CREATE TABLE artist_record (
    artist_id INTEGER REFERENCES artists(id) ON DELETE CASCADE,
    record_id INTEGER REFERENCES records(id) ON DELETE CASCADE,
    role VARCHAR(100),
    PRIMARY KEY (artist_id, record_id)
);

CREATE TABLE genre_record (
    genre_id INTEGER REFERENCES genres(id) ON DELETE CASCADE,
    record_id INTEGER REFERENCES records(id) ON DELETE CASCADE,
    PRIMARY KEY (genre_id, record_id)
);