CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    group VARCHAR(100),
    song VARCHAR(100),
    release_date DATE,
    text TEXT,
    link VARCHAR(250)
);