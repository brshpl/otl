CREATE TABLE IF NOT EXISTS links(
    id      serial PRIMARY KEY,
    data    TEXT,
    link    TEXT,
    expired BOOLEAN
);