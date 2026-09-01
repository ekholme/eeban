CREATE TABLE boards (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE columns (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    position  INTEGER NOT NULL,
    wip_limit INTEGER
);

CREATE INDEX idx_columns_board ON columns(board_id, position);

CREATE TABLE cards (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    column_id   INTEGER NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL,
    priority    INTEGER NOT NULL DEFAULT 0,
    due_date    TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    archived_at TEXT
);

CREATE INDEX idx_cards_column ON cards(column_id, position);

-- Seed a default board so the TUI always has something to render.
INSERT INTO boards (id, name) VALUES (1, 'My Board');

INSERT INTO columns (board_id, name, position) VALUES
    (1, 'Backlog',     1000),
    (1, 'In Progress', 2000),
    (1, 'Done',        3000);

INSERT INTO cards (column_id, title, body, position) VALUES
    (1, 'Welcome to eeban',    'Edit or delete this card once mutations land.', 1000),
    (1, 'Press ? for help',    '',                                             2000),
    (2, 'Build the card form', '',                                             1000),
    (3, 'Scaffold the repo',   '',                                             1000);
