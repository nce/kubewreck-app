-- Run once when provisioning the database. Safe to re-run without resetting the name.
CREATE TABLE IF NOT EXISTS teams (
    id integer PRIMARY KEY CHECK (id = 1),
    name text NOT NULL CHECK (length(trim(name)) > 0)
);

INSERT INTO teams (id, name) VALUES (1, 'opdt')
ON CONFLICT (id) DO NOTHING;
