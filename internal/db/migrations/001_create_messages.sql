CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    from_client TEXT,
    to_client TEXT,
    channel TEXT,
    content TEXT,
    timestamp DATETIME
);