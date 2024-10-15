package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Video struct {
	ID                   int64     `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	ThumbURL             string    `json:"thumb_url" db:"thumb_url"`
	ImageURL             string    `json:"image_url" db:"image_url"`
	VideoURL             string    `json:"video_url" db:"video_url"`
	SubtitlesURL         string    `json:"subtitles_url" db:"subtitles_url"`
	Description          string    `json:"description" db:"description"`
	ReleaseDate          time.Time `json:"release_date" db:"release_date"`
	Width                int       `json:"width" db:"width"`
	Height               int       `json:"height" db:"height"`
	Duration             int       `json:"duration" db:"duration"`
	Sequence             int       `json:"sequence" db:"sequence"`
	File                 string    `json:"file" db:"file"`
	OriginalFile         string    `json:"original_file" db:"original_file"`
	Path                 string    `json:"path" db:"path"`
	Md5sum               string    `json:"md5sum" db:"md5sum"`
	EnableSemanticSearch bool      `json:"enable_semantic_search" db:"enable_semantic_search"`
	Version              int32     `json:"version" db:"version"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	ModifiedAt           time.Time `json:"modified_at" db:"modified_at"`
}

type VideoModel struct {
	DB *sql.DB
}

func (m VideoModel) Insert(video *Video) error {
	query := `
		INSERT INTO videos (
			name,
			thumb_url,
			image_url,
			video_url,
			subtitles_url,
			description,
			release_date,
			width,
			height,
			duration,
			sequence,
			file,
			original_file,
			path,
			md5sum
			, enable_semantic_search
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15
			, $16
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		video.Name,
		video.ThumbURL,
		video.ImageURL,
		video.VideoURL,
		video.SubtitlesURL,
		video.Description,
		video.ReleaseDate,
		video.Width,
		video.Height,
		video.Duration,
		video.Sequence,
		video.File,
		video.OriginalFile,
		video.Path,
		video.Md5sum,
		video.EnableSemanticSearch,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&video.ID, &video.Version, &video.CreatedAt, &video.ModifiedAt)

	if err != nil {
		return videoCustomError(err)
	}

	return nil

}

func (m VideoModel) Get(id int64) (*Video, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		name,
		thumb_url,
		image_url,
		video_url,
		subtitles_url,
		description,
		release_date,
		width,
		height,
		duration,
		sequence,
		file,
		original_file,
		path,
		md5sum,
		enable_semantic_search,
		version, created_at, modified_at
		FROM videos
		WHERE id = $1`

	var video Video

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&video.ID,
		&video.Name,
		&video.ThumbURL,
		&video.ImageURL,
		&video.VideoURL,
		&video.SubtitlesURL,
		&video.Description,
		&video.ReleaseDate,
		&video.Width,
		&video.Height,
		&video.Duration,
		&video.Sequence,
		&video.File,
		&video.OriginalFile,
		&video.Path,
		&video.Md5sum,
		&video.EnableSemanticSearch,
		&video.Version,
		&video.CreatedAt,
		&video.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &video, nil
}

func (m VideoModel) Update(video *Video) error {
	query := `
		UPDATE videos
		SET
		name = $1,
		thumb_url = $2,
		image_url = $3,
		video_url = $4,
		subtitles_url = $5,
		description = $6,
		release_date = $7,
		width = $8,
		height = $9,
		duration = $10,
		sequence = $11,
		file = $12,
		original_file = $13,
		path = $14,
		md5sum = $15,
		enable_semantic_search = $16,
		version = version + 1
		WHERE id = $17 AND version = $18
		RETURNING version`

	args := []any{
		video.Name,
		video.ThumbURL,
		video.ImageURL,
		video.VideoURL,
		video.SubtitlesURL,
		video.Description,
		video.ReleaseDate,
		video.Width,
		video.Height,
		video.Duration,
		video.Sequence,
		video.File,
		video.OriginalFile,
		video.Path,
		video.Md5sum,
		video.EnableSemanticSearch,
		video.ID,
		video.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&video.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return videoCustomError(err)
		}
	}

	return nil
}

func (m VideoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM videos
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

func (m VideoModel) GetAll(name string, file string, original_file string, md5sum string, filters Filters) ([]*Video, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		thumb_url,
		image_url,
		video_url,
		subtitles_url,
		description,
		release_date,
		width,
		height,
		duration,
		sequence,
		file,
		original_file,
		path,
		md5sum,
		enable_semantic_search,
		version, created_at, modified_at
		FROM videos
		WHERE
		(to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '') AND 
		(to_tsvector('simple', file) @@ plainto_tsquery('simple', $2) OR $2 = '') AND 
		(to_tsvector('simple', original_file) @@ plainto_tsquery('simple', $3) OR $3 = '') AND 
		(to_tsvector('simple', md5sum) @@ plainto_tsquery('simple', $4) OR $4 = '')
		ORDER BY %s %s, id ASC
		LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		file,
		original_file,
		md5sum,
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
			&video.Name,
			&video.ThumbURL,
			&video.ImageURL,
			&video.VideoURL,
			&video.SubtitlesURL,
			&video.Description,
			&video.ReleaseDate,
			&video.Width,
			&video.Height,
			&video.Duration,
			&video.Sequence,
			&video.File,
			&video.OriginalFile,
			&video.Path,
			&video.Md5sum,
			&video.EnableSemanticSearch,
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

/*not_in_semantic_start*/
func (m VideoModel) GetAllNotInSemantic(filters Filters, contentField string) ([]*Video, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		description,
		version, created_at, modified_at
		FROM videos
		WHERE
		enable_semantic_search = true AND
		%s != '' AND
		id NOT IN (SELECT video_id FROM documents WHERE content_field = '%s')
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
	videos := []*Video{}

	for rows.Next() {
		var video Video

		err := rows.Scan(
			&totalRecords,
			&video.ID,
			&video.Name,
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

/*not_in_semantic_end*/
