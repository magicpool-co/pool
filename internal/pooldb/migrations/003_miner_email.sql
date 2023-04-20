ALTER TABLE miners ADD COLUMN email VARCHAR(100) AFTER address;
ALTER TABLE miners ADD COLUMN enabled_worker_notifications bool NOT NULL DEFAULT FALSE AFTER active;
ALTER TABLE miners ADD COLUMN enabled_payout_notifications bool NOT NULL DEFAULT FALSE AFTER enabled_worker_notifications;
