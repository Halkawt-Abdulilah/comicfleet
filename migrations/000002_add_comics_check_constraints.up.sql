ALTER TABLE comics ADD CONSTRAINT comics_volumes_check CHECK (volumes >= 0);
ALTER TABLE comics ADD CONSTRAINT comics_year_check CHECK (year BETWEEN 1888 AND date_part('year', now()));
ALTER TABLE comics ADD CONSTRAINT genres_length_check CHECK (array_length(genres, 1) BETWEEN 1 AND 5);