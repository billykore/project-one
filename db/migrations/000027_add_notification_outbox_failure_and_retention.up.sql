-- Messages that exhaust their delivery attempts are parked with failed_at set
-- instead of being retried forever. They stay in the table for an operator to
-- inspect and requeue (see deployments/README.md).
ALTER TABLE notification_outbox ADD COLUMN failed_at TIMESTAMPTZ;

-- The dispatcher only scans deliverable rows.
DROP INDEX idx_notification_outbox_dispatch;
CREATE INDEX idx_notification_outbox_dispatch
    ON notification_outbox (available_at, id)
    WHERE published_at IS NULL AND failed_at IS NULL;

-- Retention cleanup deletes published rows oldest first.
CREATE INDEX idx_notification_outbox_published
    ON notification_outbox (published_at)
    WHERE published_at IS NOT NULL;
