ALTER TABLE notifications ADD COLUMN event_id VARCHAR(128) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_notifications_event_id ON notifications(event_id) WHERE event_id != '';
