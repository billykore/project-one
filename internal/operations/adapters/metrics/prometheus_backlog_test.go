package metrics

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/billykore/project-one/internal/operations/domain"
)

type fakeBacklogReader struct {
	backlog     domain.DeliveryBacklog
	err         error
	gotDeadline bool
}

func (f *fakeBacklogReader) DeliveryBacklog(ctx context.Context) (domain.DeliveryBacklog, error) {
	_, f.gotDeadline = ctx.Deadline()
	return f.backlog, f.err
}

func TestPrometheusExportsDeliveryBacklogWithoutLabels(t *testing.T) {
	recorder := NewPrometheus()
	reader := &fakeBacklogReader{backlog: domain.DeliveryBacklog{Pending: 7, OldestPendingAge: 90 * time.Second, Failed: 2}}
	recorder.RegisterDeliveryBacklog(reader, 0)

	samples := gather(t, recorder.Gatherer())

	for key, want := range map[string]float64{
		`projectone_outbox_pending_messages{}`:           7,
		`projectone_outbox_oldest_pending_age_seconds{}`: 90,
		`projectone_outbox_failed_messages{}`:            2,
	} {
		if got := requireSample(t, samples, key).value; got != want {
			t.Fatalf("%s = %v, want %v", key, got, want)
		}
	}
	if !reader.gotDeadline {
		t.Fatal("backlog read had no deadline; a slow database could stall the scrape")
	}
}

func TestPrometheusOmitsDeliveryBacklogWhenReadFails(t *testing.T) {
	recorder := NewPrometheus()
	recorder.ObserveRequest(context.Background(), "GET", "/healthz", "2xx")
	recorder.RegisterDeliveryBacklog(&fakeBacklogReader{err: errors.New("database unavailable")}, time.Second)

	// The scrape itself still succeeds and keeps every other series.
	samples := gather(t, recorder.Gatherer())

	requireSample(t, samples, `projectone_http_requests_total{method="GET",route="/healthz",status_class="2xx"}`)
	for key := range samples {
		if strings.HasPrefix(key, "projectone_outbox_") {
			t.Fatalf("series %s exported although the backlog read failed", key)
		}
	}
}
