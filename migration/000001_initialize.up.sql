CREATE TABLE IF NOT EXISTS event(
    project_id UUID,
    actor_pid TEXT,
    session INTEGER,
    index INTEGER,
    PRIMARY KEY (project_id, actor_pid, session, index)
);
