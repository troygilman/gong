CREATE TABLE IF NOT EXISTS events(
    project_id UUID,
    timestamp TIMESTAMPTZ,
	id TEXT,
    pid TEXT NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    sender TEXT,
    msg TEXT,
    attrs JSONB,
    PRIMARY KEY (project_id, id, timestamp)
);

SELECT create_hypertable('events', by_range('timestamp'));

CREATE INDEX IF NOT EXISTS events_sender_project_id_pid_timestamp_idx ON events(sender, project_id, pid, timestamp);
CREATE INDEX IF NOT EXISTS events_pid_project_id_sender_timestamp_idx ON events(pid, project_id, sender, timestamp);


CREATE TABLE IF NOT EXISTS actors(
    project_id UUID,
    pid TEXT,
    num_events INTEGER NOT NULL,
    last_event_id TEXT NOT NULL,
    alive BOOLEAN NOT NULL,
    PRIMARY KEY (project_id, pid)
);