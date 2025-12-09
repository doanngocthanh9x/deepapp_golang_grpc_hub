-- Workers table
CREATE TABLE IF NOT EXISTS workers (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'online',
    metadata TEXT, -- JSON
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Capabilities table
CREATE TABLE IF NOT EXISTS capabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    worker_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    input_schema TEXT, -- JSON schema
    output_schema TEXT, -- JSON schema
    http_method TEXT DEFAULT 'POST', -- GET, POST, PUT, DELETE
    accepts_file BOOLEAN DEFAULT 0,
    file_field_name TEXT, -- Field name for file upload
    FOREIGN KEY (worker_id) REFERENCES workers(id) ON DELETE CASCADE,
    UNIQUE(worker_id, name)
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_capabilities_name ON capabilities(name);
CREATE INDEX IF NOT EXISTS idx_capabilities_worker ON capabilities(worker_id);
CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
