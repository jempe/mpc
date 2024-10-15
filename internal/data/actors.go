package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Actor struct {
	ID         int64     `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Gender     string    `json:"gender" db:"gender"`
	BirthDate  time.Time `json:"birth_date" db:"birth_date"`
	BirthPlace string    `json:"birth_place" db:"birth_place"`
	Biography  string    `json:"biography" db:"biography"`
	Height     int       `json:"height" db:"height"`
	ImageURL   string    `json:"image_url" db:"image_url"`
	Version    int32     `json:"version" db:"version"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ModifiedAt time.Time `json:"modified_at" db:"modified_at"`
}

type ActorModel struct {
	DB *sql.DB
}

func (m ActorModel) Insert(actor *Actor) error {
	query := `
		INSERT INTO actors (
			name,
			gender,
			birth_date,
			birth_place,
			biography,
			height,
			image_url
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		actor.Name,
		actor.Gender,
		actor.BirthDate,
		actor.BirthPlace,
		actor.Biography,
		actor.Height,
		actor.ImageURL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&actor.ID, &actor.Version, &actor.CreatedAt, &actor.ModifiedAt)

	if err != nil {
		return actorCustomError(err)
	}

	return nil

}

func (m ActorModel) Get(id int64) (*Actor, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		name,
		gender,
		birth_date,
		birth_place,
		biography,
		height,
		image_url,
		version, created_at, modified_at
		FROM actors
		WHERE id = $1`

	var actor Actor

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&actor.ID,
		&actor.Name,
		&actor.Gender,
		&actor.BirthDate,
		&actor.BirthPlace,
		&actor.Biography,
		&actor.Height,
		&actor.ImageURL,
		&actor.Version,
		&actor.CreatedAt,
		&actor.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &actor, nil
}

func (m ActorModel) Update(actor *Actor) error {
	query := `
		UPDATE actors
		SET
		name = $1,
		gender = $2,
		birth_date = $3,
		birth_place = $4,
		biography = $5,
		height = $6,
		image_url = $7,
		version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING version`

	args := []any{
		actor.Name,
		actor.Gender,
		actor.BirthDate,
		actor.BirthPlace,
		actor.Biography,
		actor.Height,
		actor.ImageURL,
		actor.ID,
		actor.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&actor.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return actorCustomError(err)
		}
	}

	return nil
}

func (m ActorModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM actors
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m ActorModel) GetAll(name string, gender string, filters Filters) ([]*Actor, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		gender,
		birth_date,
		birth_place,
		biography,
		height,
		image_url,
		version, created_at, modified_at
		FROM actors
		WHERE
		(to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '') AND 
		(to_tsvector('simple', gender) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		gender,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	actors := []*Actor{}

	for rows.Next() {
		var actor Actor

		err := rows.Scan(
			&totalRecords,
			&actor.ID,
			&actor.Name,
			&actor.Gender,
			&actor.BirthDate,
			&actor.BirthPlace,
			&actor.Biography,
			&actor.Height,
			&actor.ImageURL,
			&actor.Version,
			&actor.CreatedAt,
			&actor.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		actors = append(actors, &actor)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return actors, metadata, nil
}
