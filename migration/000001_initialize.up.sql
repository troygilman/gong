CREATE TABLE IF NOT EXISTS event(
	id UUID,
    project_id UUID,
    actor_pid TEXT NOT NULL,
    session INTEGER NOT NULL,
    index INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    msg TEXT NOT NULL,
    PRIMARY KEY (id, project_id)
);
