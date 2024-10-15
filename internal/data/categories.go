package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Category struct {
	ID                   int64     `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	EnableSemanticSearch bool      `json:"enable_semantic_search" db:"enable_semantic_search"`
	Version              int32     `json:"version" db:"version"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	ModifiedAt           time.Time `json:"modified_at" db:"modified_at"`
}

type CategoryModel struct {
	DB *sql.DB
}

func (m CategoryModel) Insert(category *Category) error {
	query := `
		INSERT INTO categories (
			name
			, enable_semantic_search
		)
		VALUES (
			$1
			, $2
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		category.Name,
		category.EnableSemanticSearch,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&category.ID, &category.Version, &category.CreatedAt, &category.ModifiedAt)

	if err != nil {
		return categoryCustomError(err)
	}

	return nil

}

func (m CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		name,
		enable_semantic_search,
		version, created_at, modified_at
		FROM categories
		WHERE id = $1`

	var category Category

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.EnableSemanticSearch,
		&category.Version,
		&category.CreatedAt,
		&category.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &category, nil
}

func (m CategoryModel) Update(category *Category) error {
	query := `
		UPDATE categories
		SET
		name = $1,
		enable_semantic_search = $2,
		version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version`

	args := []any{
		category.Name,
		category.EnableSemanticSearch,
		category.ID,
		category.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&category.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return categoryCustomError(err)
		}
	}

	return nil
}

func (m CategoryModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM categories
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

func (m CategoryModel) GetAll(name string, filters Filters) ([]*Category, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		enable_semantic_search,
		version, created_at, modified_at
		FROM categories
		WHERE
		(to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	categories := []*Category{}

	for rows.Next() {
		var category Category

		err := rows.Scan(
			&totalRecords,
			&category.ID,
			&category.Name,
			&category.EnableSemanticSearch,
			&category.Version,
			&category.CreatedAt,
			&category.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return categories, metadata, nil
}

/*not_in_semantic_start*/
func (m CategoryModel) GetAllNotInSemantic(filters Filters, contentField string) ([]*Category, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		version, created_at, modified_at
		FROM categories
		WHERE
		enable_semantic_search = true AND
		%s != '' AND
		id NOT IN (SELECT category_id FROM documents WHERE content_field = '%s')
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2`, contentField, contentField, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	categories := []*Category{}

	for rows.Next() {
		var category Category

		err := rows.Scan(
			&totalRecords,
			&category.ID,
			&category.Name,
			&category.Version,
			&category.CreatedAt,
			&category.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return categories, metadata, nil
}

/*not_in_semantic_end*/
