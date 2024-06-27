CREATE TABLE IF NOT EXISTS events(
    project_id TEXT,
    timestamp INTEGER,
	id TEXT,
    pid TEXT NOT NULL,
    type TEXT NOT NULL,
    level INTEGER NOT NULL,
    sender TEXT,
    msg TEXT,
    attrs TEXT,
    PRIMARY KEY (project_id, id)
) strict;

CREATE INDEX IF NOT EXISTS events_sender_project_id_pid_timestamp_idx ON events(sender, project_id, pid, timestamp);
CREATE INDEX IF NOT EXISTS events_pid_project_id_sender_timestamp_idx ON events(pid, project_id, sender, timestamp);

CREATE TABLE IF NOT EXISTS actors(
    project_id TEXT,
    pid TEXT,
    num_events INTEGER NOT NULL,
    last_event_id TEXT NOT NULL,
    alive INTEGER NOT NULL,
    PRIMARY KEY (project_id, pid)
) strict;