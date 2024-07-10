CREATE TABLE IF NOT EXISTS 'events' (
  project_id SYMBOL INDEX,
  timestamp TIMESTAMP,
  pid SYMBOL INDEX,
  sender SYMBOL INDEX,
  type SYMBOL INDEX,
  level SYMBOL INDEX,
  msg VARCHAR,
  attrs VARCHAR
) TIMESTAMP(timestamp) PARTITION BY day;

CREATE TABLE IF NOT EXISTS 'actors' (
  project_id STRING,
  pid STRING,
  num_events LONG,
  alive BOOLEAN,
  last_event_id STRING
);
