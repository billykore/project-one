DROP INDEX IF EXISTS idx_notification_outbox_published;

DROP INDEX idx_notification_outbox_dispatch;
CREATE INDEX idx_notification_outbox_dispatch
    ON notification_outbox (available_at, id)
    WHERE published_at IS NULL;

ALTER TABLE notification_outbox DROP COLUMN failed_at;
