CREATE TABLE IF NOT EXISTS event(
    project_id UUID,
    actor_pid TEXT,
    session INTEGER,
    index INTEGER,
    timestamp TIMESTAMP NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    msg TEXT NOT NULL,
    PRIMARY KEY (project_id, actor_pid, session, index)
);
