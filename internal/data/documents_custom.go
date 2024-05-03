package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pgvector/pgvector-go"
)

func (m DocumentModel) UpdateEmbedding(Document *Document, embeddings pgvector.Vector, provider string) error {
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
		Document.ID,
		Document.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&Document.Version)
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

func (m DocumentModel) GetAllSemantic(embedding pgvector.Vector, similarity float64, provider string, content_field string, video_id int64, filters Filters) ([]*Document, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
			content,
			tokens,
			sequence,
			content_field,
			video_id,
			version, created_at, modified_at, 1 - (%s <=> $3) AS cosine_similarity
		FROM documents
		WHERE
		(content_field = $1 OR $1 = '') AND
		(video_id = $2 OR $2 = 0) AND
		1 - (%s <=> $3) > $4
		ORDER BY cosine_similarity DESC
		LIMIT $5 OFFSET $6`, embeddings_field, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		content_field,
		video_id,
		embedding,
		similarity,
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
			&Document.Similarity,
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

func (m DocumentModel) GetAllWithoutEmbeddings(limit int, provider string) ([]*Document, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
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
	Documents := []*Document{}

	for rows.Next() {
		var Document Document

		err := rows.Scan(
			&totalRecords,
			&Document.ID,
			&Document.Content,
			&Document.Version,
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

//custom_code
