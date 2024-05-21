CREATE TABLE IF NOT EXISTS event(
	id VARCHAR(26),
    project_id UUID,
    actor_pid TEXT NOT NULL,
    index INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    sender TEXT NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    msg TEXT NOT NULL,
    attrs JSONB NOT NULL,
    PRIMARY KEY (id, project_id)
);

CREATE INDEX IF NOT EXISTS index_actor_pid ON event(actor_pid);

CREATE TABLE IF NOT EXISTS actor(
    pid TEXT,
    project_id UUID,
    num_events INTEGER NOT NULL,
    last_event_id VARCHAR(26) NOT NULL,
    alive BOOLEAN NOT NULL,
    PRIMARY KEY (pid, project_id)
);