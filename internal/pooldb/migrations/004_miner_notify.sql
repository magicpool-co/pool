ALTER TABLE workers ADD COLUMN notified bool NOT NULL DEFAULT FALSE AFTER active;
