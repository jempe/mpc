package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Document struct {
	ID           int64     `json:"id" db:"id"`
	VideoID      int64     `json:"video_id" db:"video_id"`
	CategoryID   int64     `json:"category_id" db:"category_id"`
	Title        string    `json:"title" db:"title"`
	Similarity   float64   `json:"similarity" db:"similarity"`
	Content      string    `json:"content" db:"content"`
	Tokens       int       `json:"tokens" db:"tokens"`
	Sequence     int       `json:"sequence" db:"sequence"`
	ContentField string    `json:"content_field" db:"content_field"`
	Version      int32     `json:"version" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ModifiedAt   time.Time `json:"modified_at" db:"modified_at"`
}

type DocumentModel struct {
	DB *sql.DB
}

func (m DocumentModel) Insert(document *Document) error {
	query := `
		INSERT INTO documents (
			video_id,
			category_id
			, title, content, tokens, sequence, content_field
		)
		VALUES (
			$1,
			$2
			, $3, $4, $5, $6, $7
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		document.VideoID,
		document.CategoryID,
		document.Title,
		document.Content,
		document.Tokens,
		document.Sequence,
		document.ContentField,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&document.ID, &document.Version, &document.CreatedAt, &document.ModifiedAt)

	if err != nil {
		return documentCustomError(err)
	}

	return nil

}

func (m DocumentModel) Get(id int64) (*Document, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		video_id,
		category_id,
		title, content, tokens, sequence, content_field,
		version, created_at, modified_at
		FROM documents
		WHERE id = $1`

	var document Document

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&document.ID,
		&document.VideoID,
		&document.CategoryID,
		&document.Title,
		&document.Content,
		&document.Tokens,
		&document.Sequence,
		&document.ContentField,
		&document.Version,
		&document.CreatedAt,
		&document.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &document, nil
}

func (m DocumentModel) Update(document *Document) error {
	query := `
		UPDATE documents
		SET
		video_id = $1,
		category_id = $2,
		title = $3,
		content = $4,
		tokens = $5,
		sequence = $6,
		content_field = $7,
		version = version + 1
		WHERE id = $8 AND version = $9
		RETURNING version`

	args := []any{
		document.VideoID,
		document.CategoryID,
		document.Title,
		document.Content,
		document.Tokens,
		document.Sequence,
		document.ContentField,
		document.ID,
		document.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&document.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return documentCustomError(err)
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

func (m DocumentModel) GetAll(video_id int64, category_id int64, filters Filters) ([]*Document, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		video_id,
		category_id,
		title, content, tokens, sequence, content_field,
		version, created_at, modified_at
		FROM documents
		WHERE
		(video_id = $1 OR $1 = 0) AND 
		(category_id = $2 OR $2 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		video_id,
		category_id,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	documents := []*Document{}

	for rows.Next() {
		var document Document

		err := rows.Scan(
			&totalRecords,
			&document.ID,
			&document.VideoID,
			&document.CategoryID,
			&document.Title,
			&document.Content,
			&document.Tokens,
			&document.Sequence,
			&document.ContentField,
			&document.Version,
			&document.CreatedAt,
			&document.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return documents, metadata, nil
}

// update_embedding_start
func (m DocumentModel) UpdateEmbedding(document *Document, embeddings pgvector.Vector, provider string) error {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE documents
		SET
		%s = $1,
		version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`, embeddings_field)

	args := []any{
		embeddings,
		document.ID,
		document.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&document.Version)
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

//update_embedding_end

/*get_all_semantic_start*/
func (m DocumentModel) GetAllSemantic(embedding pgvector.Vector, similarity float64, provider string, content_fields []string, video_id int64, category_id int64, filters Filters) ([]*Document, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), documents.id,
			documents.video_id,
			documents.category_id,
			documents.title,
			documents.content, 
			documents.tokens, 
			documents.sequence, 
			documents.content_field,
			documents.version, documents.created_at, documents.modified_at, 1 - (%s <=> $3) AS cosine_similarity
		FROM documents
		WHERE
		(documents.video_id = $1 OR $1 = 0) AND
		(documents.category_id = $2 OR $2 = 0) AND
		1 - (%s <=> $3) > $4 AND
		(documents.content_field = ANY ($5) OR $5 = '{}')
		ORDER BY cosine_similarity DESC
		LIMIT $6 OFFSET $7`, embeddings_field, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		video_id,
		category_id,
		embedding,
		similarity,
		pq.Array(content_fields),
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	documents := []*Document{}

	for rows.Next() {
		var document Document

		err := rows.Scan(
			&totalRecords,
			&document.ID,
			&document.VideoID,
			&document.CategoryID,
			&document.Title,
			&document.Content,
			&document.Tokens,
			&document.Sequence,
			&document.ContentField,
			&document.Version,
			&document.CreatedAt,
			&document.ModifiedAt,
			&document.Similarity,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return documents, metadata, nil
}

/*get_all_semantic_end*/

/*get_all_without_embeddings_start*/
func (m DocumentModel) GetAllWithoutEmbeddings(limit int, provider string) ([]*Document, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		title,
		content,
		version
		FROM documents
		WHERE %s IS NULL
		AND content != ''
		LIMIT $1
		`, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		limit,
	}

	filters := Filters{
		Page:     1,
		PageSize: limit,
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	documents := []*Document{}

	for rows.Next() {
		var document Document

		err := rows.Scan(
			&totalRecords,
			&document.ID,
			&document.Title,
			&document.Content,
			&document.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return documents, metadata, nil
}

/*get_all_without_embeddings_end*/
