CREATE TABLE IF NOT EXISTS event(
    project_id UUID,
    timestamp TIMESTAMPTZ,
	id VARCHAR(26),
    actor_pid TEXT NOT NULL,
    index INTEGER NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    sender TEXT,
    msg TEXT,
    attrs JSONB,
    PRIMARY KEY (project_id, id, timestamp)
);

SELECT create_hypertable('event', by_range('timestamp'));

CREATE INDEX IF NOT EXISTS event_sender_project_id_actor_pid_timestamp_idx ON event(sender, project_id, actor_pid, timestamp);
CREATE INDEX IF NOT EXISTS event_actor_pid_project_id_sender_timestamp_idx ON event(actor_pid, project_id, sender, timestamp);


CREATE TABLE IF NOT EXISTS actor(
    pid TEXT,
    project_id UUID,
    num_events INTEGER NOT NULL,
    last_event_id VARCHAR(26) NOT NULL,
    alive BOOLEAN NOT NULL,
    PRIMARY KEY (pid, project_id)
);