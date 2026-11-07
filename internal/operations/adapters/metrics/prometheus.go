// Package metrics adapts the application to the Prometheus exposition format.
// Metric labels are bounded: no account identifier, token, request identifier,
// request content, query string, or raw error ever becomes a label value.
package metrics

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/billykore/project-one/internal/operations/ports"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "projectone"

// Prometheus records bounded operational measurements and serves the
// authenticated scrape endpoint.
type Prometheus struct {
	registry     *prometheus.Registry
	requests     *prometheus.CounterVec
	durations    *prometheus.HistogramVec
	dependencyUp *prometheus.GaugeVec
	ready        prometheus.Gauge
}

// NewPrometheus creates the application registry with its bounded-label metrics
// plus the standard Go runtime and process collectors, which already export
// process_start_time_seconds for restart detection.
func NewPrometheus() *Prometheus {
	registry := prometheus.NewRegistry()

	recorder := &Prometheus{
		registry: registry,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Completed application requests grouped by method, route template, and status class.",
		}, []string{"method", "route", "status_class"}),
		durations: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Completed application request duration in seconds grouped by method and route template.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route"}),
		dependencyUp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "dependency_up",
			Help:      "Latest readiness check result per required dependency: 1 when up, otherwise 0.",
		}, []string{"dependency"}),
		ready: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ready",
			Help:      "Whether the latest readiness assessment was ready to serve traffic: 1 when ready, otherwise 0.",
		}),
	}

	registry.MustRegister(
		recorder.requests,
		recorder.durations,
		recorder.dependencyUp,
		recorder.ready,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return recorder
}

// ObserveRequest records one completed request. route must be a resolved route
// template or the single unmatched placeholder.
func (p *Prometheus) ObserveRequest(_ context.Context, method, route, statusClass string) {
	p.requests.WithLabelValues(method, route, statusClass).Inc()
}

// ObserveRequestDuration records the duration of one completed request.
func (p *Prometheus) ObserveRequestDuration(_ context.Context, method, route string, seconds float64) {
	p.durations.WithLabelValues(method, route).Observe(seconds)
}

// ObserveAssessment updates readiness and per-dependency gauges. Only the
// component name and its state are used; no diagnostic detail is recorded.
func (p *Prometheus) ObserveAssessment(_ context.Context, assessment domain.HealthAssessment) {
	if assessment.Ready() {
		p.ready.Set(1)
	} else {
		p.ready.Set(0)
	}

	for _, component := range assessment.Components {
		up := 0.0
		if component.Status == domain.ComponentStatusUp {
			up = 1
		}
		p.dependencyUp.WithLabelValues(component.Name).Set(up)
	}
}

// DefaultBacklogTimeout bounds the backlog query run on every scrape.
const DefaultBacklogTimeout = 2 * time.Second

// RegisterDeliveryBacklog exports the event delivery backlog, read from reader
// on every scrape. A failed or slow read omits the backlog series for that
// scrape instead of failing it, so the other metrics stay available and the
// gap is detectable with absent(). A timeout of zero or less uses
// DefaultBacklogTimeout.
func (p *Prometheus) RegisterDeliveryBacklog(reader ports.DeliveryBacklogReader, timeout time.Duration) {
	if timeout <= 0 {
		timeout = DefaultBacklogTimeout
	}
	p.registry.MustRegister(newBacklogCollector(reader, timeout))
}

type backlogCollector struct {
	reader    ports.DeliveryBacklogReader
	timeout   time.Duration
	pending   *prometheus.Desc
	oldestAge *prometheus.Desc
	failed    *prometheus.Desc
}

func newBacklogCollector(reader ports.DeliveryBacklogReader, timeout time.Duration) *backlogCollector {
	return &backlogCollector{
		reader:  reader,
		timeout: timeout,
		pending: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "outbox", "pending_messages"),
			"Events waiting in the transactional outbox for delivery or retry.",
			nil, nil,
		),
		oldestAge: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "outbox", "oldest_pending_age_seconds"),
			"Age in seconds of the oldest event waiting in the outbox, or 0 when none is pending.",
			nil, nil,
		),
		failed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "outbox", "failed_messages"),
			"Events parked in the outbox after exhausting their delivery attempts.",
			nil, nil,
		),
	}
}

func (c *backlogCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.pending
	ch <- c.oldestAge
	ch <- c.failed
}

func (c *backlogCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	backlog, err := c.reader.DeliveryBacklog(ctx)
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(c.pending, prometheus.GaugeValue, float64(backlog.Pending))
	ch <- prometheus.MustNewConstMetric(c.oldestAge, prometheus.GaugeValue, backlog.OldestPendingAge.Seconds())
	ch <- prometheus.MustNewConstMetric(c.failed, prometheus.GaugeValue, float64(backlog.Failed))
}

// Gatherer exposes the registry for tests.
func (p *Prometheus) Gatherer() prometheus.Gatherer { return p.registry }

// Handler returns the Prometheus scrape handler. The credential is mandatory:
// a nil credential or an invalid one yields 401 with no metric content.
func (p *Prometheus) Handler(credential *Credential) http.Handler {
	return requireCredential(promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}), credential)
}

// Credential is the dedicated machine identity used to scrape metrics.
// It is not an application account and is never logged or returned.
type Credential struct {
	Username string
	Password string
}

// LoadCredential reads the monitoring secret from its mounted file. An empty
// username and path mean monitoring is disabled and returns (nil, nil), which
// makes the scrape endpoint reject every request.
func LoadCredential(username, passwordFile string) (*Credential, error) {
	if username == "" && passwordFile == "" {
		return nil, nil
	}
	if username == "" || passwordFile == "" {
		return nil, errors.New("monitoring username and password file must be configured together")
	}

	secret, err := os.ReadFile(passwordFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read monitoring password file: %w", err)
	}
	password := strings.TrimSpace(string(secret))
	if password == "" {
		return nil, fmt.Errorf("monitoring password file %s is empty", passwordFile)
	}

	return &Credential{Username: username, Password: password}, nil
}

func requireCredential(next http.Handler, credential *Credential) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthorized(r, credential) {
			w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAuthorized(r *http.Request, credential *Credential) bool {
	if credential == nil || credential.Username == "" {
		return false
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// Constant-time comparison keeps the credential out of timing side channels.
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(credential.Username)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(credential.Password)) == 1

	return userMatch && passwordMatch
}

var _ ports.HTTPMetrics = (*Prometheus)(nil)
var _ ports.HealthObserver = (*Prometheus)(nil)
