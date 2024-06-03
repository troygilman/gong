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

-- may need to add projectID, timestamp to these indexes
CREATE INDEX IF NOT EXISTS event_actor_pid_sender_idx ON event(actor_pid, sender);
CREATE INDEX IF NOT EXISTS event_sender_actor_pid_idx ON event(sender, actor_pid);

CREATE INDEX IF NOT EXISTS index_type ON event(type);
CREATE INDEX IF NOT EXISTS index_level ON event(level);

CREATE TABLE IF NOT EXISTS actor(
    pid TEXT,
    project_id UUID,
    num_events INTEGER NOT NULL,
    last_event_id VARCHAR(26) NOT NULL,
    alive BOOLEAN NOT NULL,
    PRIMARY KEY (pid, project_id)
);