CREATE TABLE parents (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE children (
    id SERIAL PRIMARY KEY,
    parent_id INT NOT NULL REFERENCES parents (id),
    name TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
