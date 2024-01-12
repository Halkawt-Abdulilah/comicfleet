package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"comicfleet.halkawtabdulilah.net/internal/validator"
	"github.com/lib/pq"
)

type Comic struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year ,omitempty"`
	Volumes   Volumes   `json:"volumes ,omitempty"`
	Genres    []string  `json:"genres ,omitempty"`
	Version   int32     `json:"version"`
}

type ComicModel struct {
	DB *sql.DB
}

func (c ComicModel) Insert(comic *Comic) error {
	query := `
		INSERT INTO comics (title, year, volumes, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{comic.Title, comic.Year, comic.Volumes, pq.Array(comic.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comic.ID, &comic.CreatedAt, &comic.Version)
}

func (c ComicModel) Get(id int64) (*Comic, error) {

	//could switch to unsigned integer but not with PostgreSQL
	if id < 1 {
		return nil, RecordNotFoundError
	}

	query := `
		SELECT id, created_at, title, year, volumes, genres, version
		FROM comics
		WHERE id = $1`

	var comic Comic

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&comic.ID,
		&comic.CreatedAt,
		&comic.Title,
		&comic.Year,
		&comic.Volumes,
		pq.Array(&comic.Genres),
		&comic.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, sql.ErrNoRows
		default:
			return nil, err
		}
	}

	return &comic, nil
}

func (c ComicModel) GetAll(title string, genres []string, filters Filters) ([]*Comic, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, volumes, genres, version
		FROM comics
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}

	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	comics := []*Comic{}

	for rows.Next() {
		var comic Comic

		err := rows.Scan(
			&totalRecords,
			&comic.ID,
			&comic.CreatedAt,
			&comic.Title,
			&comic.Year,
			&comic.Volumes,
			pq.Array(&comic.Genres),
			&comic.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		comics = append(comics, &comic)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return comics, metadata, nil
}

func (c ComicModel) Update(comic *Comic) error {

	query := `
		UPDATE comics
		SET title = $1, year = $2, volumes = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		comic.Title,
		comic.Year,
		comic.Volumes,
		pq.Array(comic.Genres),
		comic.ID,
		comic.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&comic.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return EditConflictError
		default:
			return err
		}
	}

	return nil
}

func (c ComicModel) Delete(id int64) error {

	if id < 1 {
		return RecordNotFoundError
	}

	query := `
		DELETE FROM comics
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return RecordNotFoundError
	}

	return nil
}

func ValidateComic(v *validator.Validator, comic *Comic) {
	v.Check(comic.Title != "", "title", "must be provided")
	v.Check(len(comic.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(comic.Year != 0, "year", "must be provided")
	v.Check(comic.Year >= 1888, "year", "must be greater than 1888")
	v.Check(comic.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(comic.Volumes != 0, "volumes", "must be provided")
	v.Check(comic.Volumes > 0, "volumes", "must be a positive integer")

	v.Check(comic.Genres != nil, "genres", "must be provided")
	v.Check(len(comic.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(comic.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(comic.Genres), "genres", "must not contain duplicate values")
}
