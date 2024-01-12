CREATE INDEX IF NOT EXISTS comics_title_idx ON comics USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS comics_genres_idx ON comics USING GIN (genres);