CREATE TABLE IF NOT EXISTS event(
	id UUID,
    project_id UUID,
    actor_pid TEXT NOT NULL,
    index INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    sender TEXT NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    msg TEXT NOT NULL,
    PRIMARY KEY (id, project_id)
);
