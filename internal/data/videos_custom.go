package data

import (
	"context"
	"fmt"
	"time"
)

func (m VideoModel) GetAllNotInSemantic(filters Filters) ([]*Video, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		description,
		version, created_at, modified_at
		FROM videos
		WHERE
		enable_semantic_search = true AND
		id NOT IN (SELECT video_id FROM documents)
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2`, filters.sortColumn(), filters.sortDirection())

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
	videos := []*Video{}

	for rows.Next() {
		var video Video

		err := rows.Scan(
			&totalRecords,
			&video.ID,
			&video.Description,
			&video.Version,
			&video.CreatedAt,
			&video.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		videos = append(videos, &video)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return videos, metadata, nil
}

//custom_code
