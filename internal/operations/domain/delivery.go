package domain

import "time"

// DeliveryBacklog is the non-sensitive state of the event delivery backlog. It
// carries only counts and an age, never event identifiers or payloads.
type DeliveryBacklog struct {
	// Pending counts events still being delivered or retried.
	Pending int64
	// OldestPendingAge is how long the oldest pending event has waited, or
	// zero when nothing is pending.
	OldestPendingAge time.Duration
	// Failed counts events parked after exhausting their delivery attempts.
	Failed int64
}
