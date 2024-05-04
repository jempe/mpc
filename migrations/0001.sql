BEGIN;
--run the following with superuser privileges:
--CREATE EXTENSION IF NOT EXISTS citext;
--CREATE EXTENSION IF NOT EXISTS vector;

-- migration table start

CREATE TABLE IF NOT EXISTS migrations (
        version INTEGER UNIQUE NOT NULL DEFAULT 1
);

INSERT INTO migrations (version) VALUES (1);

-- migration table end

-- categories table start

CREATE TABLE IF NOT EXISTS categories (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- auto update modified_at
CREATE OR REPLACE FUNCTION categories_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER categories_trigger_update_modified_at
BEFORE UPDATE ON categories
FOR EACH ROW
EXECUTE PROCEDURE categories_update_modified_at();

-- categories table end

-- actors table start

CREATE TABLE IF NOT EXISTS actors (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	gender TEXT NOT NULL,
	birth_date DATE NOT NULL,
	birth_place TEXT NOT NULL,
	biography TEXT NOT NULL,
	height INTEGER NOT NULL DEFAULT 0,
	image_url TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- constraint for name length
ALTER TABLE actors
ADD CONSTRAINT actors_check_name_length
CHECK (LENGTH(name) BETWEEN 3 AND 200);

-- constraint for gender
ALTER TABLE actors
ADD CONSTRAINT actors_check
CHECK (gender IN ('male', 'female'));

-- auto update modified_at
CREATE OR REPLACE FUNCTION actors_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER actors_trigger_update_modified_at
BEFORE UPDATE ON actors
FOR EACH ROW
EXECUTE PROCEDURE actors_update_modified_at();

-- actors table end

-- videos table start

CREATE TABLE IF NOT EXISTS videos (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	thumb_url TEXT NOT NULL,
	image_url TEXT NOT NULL,
	video_url TEXT NOT NULL,
	subtitles_url TEXT NOT NULL,
	description TEXT NOT NULL,
	release_date DATE NOT NULL,
	width INTEGER NOT NULL DEFAULT 0,
	height INTEGER NOT NULL DEFAULT 0,
	duration INTEGER NOT NULL DEFAULT 0,
	sequence INTEGER NOT NULL DEFAULT 0,
	file TEXT NOT NULL,
	original_file TEXT NOT NULL,
	path TEXT NOT NULL,
	md5sum TEXT NOT NULL UNIQUE,
	enable_semantic_search BOOLEAN NOT NULL DEFAULT TRUE,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- constraint for title length
ALTER TABLE videos
ADD CONSTRAINT videos_check_title_length
CHECK (LENGTH(title) BETWEEN 3 AND 200);

-- auto update modified_at

CREATE OR REPLACE FUNCTION videos_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER videos_trigger_update_modified_at
BEFORE UPDATE ON videos
FOR EACH ROW
EXECUTE PROCEDURE videos_update_modified_at();


-- indexes
CREATE INDEX IF NOT EXISTS movies_title_idx ON videos USING GIN (to_tsvector('simple', title));

-- videos table end

-- documents table start

CREATE TABLE IF NOT EXISTS documents (
	id BIGSERIAL PRIMARY KEY,
	content TEXT NOT NULL,
	openai_embeddings vector(1536),
	st_embeddings vector(384),
	tokens INT NOT NULL DEFAULT 0,
	sequence INT NOT NULL DEFAULT 0,
	content_field TEXT NOT NULL,
	generic_item_id BIGINT NOT NULL REFERENCES videos(id),
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE documents
ADD CONSTRAINT chk_content_field
CHECK (content_field IN ('videos.name', 'videos.description', 'videos.content'));

-- auto update modified_at
CREATE OR REPLACE FUNCTION documents_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER documents_trigger_update_modified_at
BEFORE UPDATE ON documents
FOR EACH ROW
EXECUTE PROCEDURE documents_update_modified_at();

-- documents table end

-- videos_categories table start

CREATE TABLE IF NOT EXISTS videos_categories (
	video_id BIGINT NOT NULL REFERENCES videos(id),
	category_id BIGINT NOT NULL REFERENCES categories(id)
);

--create unique constraint
ALTER TABLE videos_categories
ADD CONSTRAINT videos_categories_unique
UNIQUE (video_id, category_id);

-- indexes
CREATE INDEX IF NOT EXISTS videos_categories_video_id_idx ON videos_categories(video_id);
CREATE INDEX IF NOT EXISTS videos_categories_category_id_idx ON videos_categories(category_id);

-- videos_categories table end

COMMIT;

