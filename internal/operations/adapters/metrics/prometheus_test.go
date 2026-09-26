package metrics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
	"github.com/prometheus/client_golang/prometheus"
)

// metricSample holds the single value of a counter or gauge, or the count and
// sum of a histogram, keyed by its full series name.
type metricSample struct {
	value float64
	count uint64
	sum   float64
}

// gather reads every series the Prometheus registry would expose.
func gather(t *testing.T, gatherer prometheus.Gatherer) map[string]metricSample {
	t.Helper()

	families, err := gatherer.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	samples := make(map[string]metricSample)
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := make([]string, 0, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels = append(labels, fmt.Sprintf("%s=%q", label.GetName(), label.GetValue()))
			}
			sort.Strings(labels)
			key := family.GetName() + "{" + strings.Join(labels, ",") + "}"

			switch {
			case metric.GetCounter() != nil:
				samples[key] = metricSample{value: metric.GetCounter().GetValue()}
			case metric.GetGauge() != nil:
				samples[key] = metricSample{value: metric.GetGauge().GetValue()}
			case metric.GetHistogram() != nil:
				histogram := metric.GetHistogram()
				samples[key] = metricSample{count: histogram.GetSampleCount(), sum: histogram.GetSampleSum()}
			}
		}
	}
	return samples
}

func requireSample(t *testing.T, samples map[string]metricSample, key string) metricSample {
	t.Helper()

	sample, ok := samples[key]
	if !ok {
		keys := make([]string, 0, len(samples))
		for candidate := range samples {
			keys = append(keys, candidate)
		}
		sort.Strings(keys)
		t.Fatalf("series %s not found among:\n%s", key, strings.Join(keys, "\n"))
	}
	return sample
}

func TestPrometheusRecordsRequestsWithBoundedLabelsOnly(t *testing.T) {
	recorder := NewPrometheus()
	ctx := context.Background()

	recorder.ObserveRequest(ctx, "GET", "/posts/:id", "4xx")
	recorder.ObserveRequest(ctx, "GET", "/posts/:id", "4xx")
	recorder.ObserveRequest(ctx, "POST", "/posts", "2xx")
	recorder.ObserveRequest(ctx, "", "", "5xx")

	samples := gather(t, recorder.Gatherer())

	repeated := requireSample(t, samples, `projectone_http_requests_total{method="GET",route="/posts/:id",status_class="4xx"}`)
	if repeated.value != 2 {
		t.Fatalf("repeat count = %v, want 2", repeated.value)
	}
	created := requireSample(t, samples, `projectone_http_requests_total{method="POST",route="/posts",status_class="2xx"}`)
	if created.value != 1 {
		t.Fatalf("create count = %v, want 1", created.value)
	}
}

func TestPrometheusRequestLabelNamesStayBounded(t *testing.T) {
	recorder := NewPrometheus()
	recorder.ObserveRequest(context.Background(), "GET", "/users/:username", "2xx")

	families, err := recorder.Gatherer().Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}

	var labelNames []string
	for _, family := range families {
		if family.GetName() != "projectone_http_requests_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				labelNames = append(labelNames, label.GetName())
			}
		}
	}
	sort.Strings(labelNames)

	want := []string{"method", "route", "status_class"}
	if strings.Join(labelNames, ",") != strings.Join(want, ",") {
		t.Fatalf("label names = %v, want %v", labelNames, want)
	}
}

func TestPrometheusRecordsRequestDurationInSeconds(t *testing.T) {
	recorder := NewPrometheus()
	ctx := context.Background()

	recorder.ObserveRequestDuration(ctx, "POST", "/posts", 0.25)
	recorder.ObserveRequestDuration(ctx, "POST", "/posts", 0.75)

	samples := gather(t, recorder.Gatherer())

	observed := requireSample(t, samples, `projectone_http_request_duration_seconds{method="POST",route="/posts"}`)
	if observed.count != 2 {
		t.Fatalf("sample count = %d, want 2", observed.count)
	}
	if observed.sum < 0.99 || observed.sum > 1.01 {
		t.Fatalf("sample sum = %v, want 1.0 within floating point tolerance", observed.sum)
	}
}

func TestPrometheusExposesReadinessGauge(t *testing.T) {
	recorder := NewPrometheus()
	ctx := context.Background()

	recorder.ObserveAssessment(ctx, domain.HealthAssessment{
		Status:     domain.ReadinessStatusReady,
		CheckedAt:  time.Now().UTC(),
		Components: []domain.ComponentHealth{{Name: "database", Status: domain.ComponentStatusUp}},
	})
	if ready := requireSample(t, gather(t, recorder.Gatherer()), "projectone_ready{}"); ready.value != 1 {
		t.Fatalf("ready = %v, want 1", ready.value)
	}

	recorder.ObserveAssessment(ctx, domain.HealthAssessment{
		Status:     domain.ReadinessStatusNotReady,
		CheckedAt:  time.Now().UTC(),
		Components: []domain.ComponentHealth{{Name: "database", Status: domain.ComponentStatusUnknown}},
	})
	if ready := requireSample(t, gather(t, recorder.Gatherer()), "projectone_ready{}"); ready.value != 0 {
		t.Fatalf("ready = %v, want 0 after a failed assessment", ready.value)
	}
}

func TestPrometheusExposesDependencyGaugePerComponent(t *testing.T) {
	recorder := NewPrometheus()
	ctx := context.Background()

	recorder.ObserveAssessment(ctx, domain.HealthAssessment{
		Status:    domain.ReadinessStatusNotReady,
		CheckedAt: time.Now().UTC(),
		Components: []domain.ComponentHealth{
			{Name: "database", Status: domain.ComponentStatusDown},
			{Name: "rabbitmq", Status: domain.ComponentStatusUp},
		},
	})

	samples := gather(t, recorder.Gatherer())
	if database := requireSample(t, samples, `projectone_dependency_up{dependency="database"}`); database.value != 0 {
		t.Fatalf("database = %v, want 0", database.value)
	}
	if broker := requireSample(t, samples, `projectone_dependency_up{dependency="rabbitmq"}`); broker.value != 1 {
		t.Fatalf("rabbitmq = %v, want 1", broker.value)
	}

	recorder.ObserveAssessment(ctx, domain.HealthAssessment{
		Status:     domain.ReadinessStatusReady,
		CheckedAt:  time.Now().UTC(),
		Components: []domain.ComponentHealth{{Name: "database", Status: domain.ComponentStatusUp}},
	})
	if database := requireSample(t, gather(t, recorder.Gatherer()), `projectone_dependency_up{dependency="database"}`); database.value != 1 {
		t.Fatalf("database = %v, want 1 after recovery", database.value)
	}
}

func TestPrometheusRegistersStandardCollectors(t *testing.T) {
	recorder := NewPrometheus()

	samples := gather(t, recorder.Gatherer())
	if _, ok := samples["process_start_time_seconds{}"]; !ok {
		t.Fatal("process_start_time_seconds is missing; an operator cannot detect a restart")
	}

	found := false
	for key := range samples {
		if strings.HasPrefix(key, "go_goroutines{") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("go_goroutines is missing; the Go runtime collector is not registered")
	}
}

func TestLoadCredentialReadsTheMountedSecretAndTrimsIt(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "metrics-password")
	if err := os.WriteFile(secretPath, []byte("mounted-secret\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	credential, err := LoadCredential("projectone-metrics", secretPath)
	if err != nil {
		t.Fatalf("LoadCredential returned %v, want nil", err)
	}
	if credential.Username != "projectone-metrics" || credential.Password != "mounted-secret" {
		t.Fatalf("credential = %+v, want the trimmed mounted secret", credential)
	}
}

func TestLoadCredentialRejectsIncompleteOrMissingSecrets(t *testing.T) {
	if credential, err := LoadCredential("", ""); err != nil || credential != nil {
		t.Fatalf("disabled monitoring = (%v, %v), want (nil, nil)", credential, err)
	}

	if _, err := LoadCredential("projectone-metrics", ""); err == nil {
		t.Fatal("expected an error for a username without a password file")
	}

	emptyPath := filepath.Join(t.TempDir(), "metrics-password")
	if err := os.WriteFile(emptyPath, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCredential("projectone-metrics", emptyPath); err == nil {
		t.Fatal("expected an error for an empty password file")
	}

	if _, err := LoadCredential("projectone-metrics", filepath.Join(t.TempDir(), "absent")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing secret error = %v, want a file-not-found error", err)
	}
}
