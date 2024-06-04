CREATE TABLE IF NOT EXISTS event(
    project_id UUID,
    timestamp TIMESTAMPTZ,
	id VARCHAR(26),
    actor_pid TEXT NOT NULL,
    index INTEGER NOT NULL,
    sender TEXT NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    msg TEXT NOT NULL,
    attrs JSONB NOT NULL,
    PRIMARY KEY (project_id, id, timestamp)
);

SELECT create_hypertable('event', by_range('timestamp'));

-- may need to add projectID, timestamp to these indexes
CREATE INDEX IF NOT EXISTS event_sources_idx ON event(project_id, actor_pid, sender);
CREATE INDEX IF NOT EXISTS event_sinks_idx ON event(project_id, sender, actor_pid);


CREATE TABLE IF NOT EXISTS actor(
    pid TEXT,
    project_id UUID,
    num_events INTEGER NOT NULL,
    last_event_id VARCHAR(26) NOT NULL,
    alive BOOLEAN NOT NULL,
    PRIMARY KEY (pid, project_id)
);