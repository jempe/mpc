package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Document struct {
	ID           int64     `json:"id,omitempty" db:"id"`
	Content      string    `json:"content,omitempty" db:"content"`
	Tokens       int       `json:"tokens,omitempty" db:"tokens"`
	Sequence     int       `json:"sequence,omitempty" db:"sequence"`
	ContentField string    `json:"content_field,omitempty" db:"content_field"`
	VideoID      int64     `json:"video_id,omitempty" db:"video_id"`
	Similarity   float64   `json:"similarity,omitempty" db:"similarity"`
	Version      int32     `json:"version,omitempty" db:"version"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
	ModifiedAt   time.Time `json:"-" db:"modified_at"`
}

type DocumentModel struct {
	DB *sql.DB
}

func (m DocumentModel) Insert(Document *Document) error {
	query := `
		INSERT INTO documents (
			content,
			tokens,
			sequence,
			content_field,
			video_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		Document.Content,
		Document.Tokens,
		Document.Sequence,
		Document.ContentField,
		Document.VideoID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&Document.ID, &Document.Version, &Document.CreatedAt, &Document.ModifiedAt)
}

func (m DocumentModel) Get(id int64) (*Document, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		content,
		tokens,
		sequence,
		content_field,
		video_id,
		version, created_at, modified_at
		FROM documents
		WHERE id = $1`

	var Document Document

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&Document.ID,
		&Document.Content,
		&Document.Tokens,
		&Document.Sequence,
		&Document.ContentField,
		&Document.VideoID,
		&Document.Version,
		&Document.CreatedAt,
		&Document.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &Document, nil
}

func (m DocumentModel) Update(Document *Document) error {
	query := `
		UPDATE documents
		SET
		content = $1,
		tokens = $2,
		sequence = $3,
		content_field = $4,
		video_id = $5,
		version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version`

	args := []any{
		Document.Content,
		Document.Tokens,
		Document.Sequence,
		Document.ContentField,
		Document.VideoID,
		Document.ID,
		Document.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&Document.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m DocumentModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM documents
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

func (m DocumentModel) GetAll(content_field string, video_id int64, filters Filters) ([]*Document, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		content,
		tokens,
		sequence,
		content_field,
		video_id,
		version, created_at, modified_at
		FROM documents
		WHERE
		(to_tsvector('simple', content_field) @@ plainto_tsquery('simple', $1) OR $1 = '') AND 
		(video_id = $2 OR $2 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		content_field,
		video_id,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	Documents := []*Document{}

	for rows.Next() {
		var Document Document

		err := rows.Scan(
			&totalRecords,
			&Document.ID,
			&Document.Content,
			&Document.Tokens,
			&Document.Sequence,
			&Document.ContentField,
			&Document.VideoID,
			&Document.Version,
			&Document.CreatedAt,
			&Document.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		Documents = append(Documents, &Document)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return Documents, metadata, nil
}
