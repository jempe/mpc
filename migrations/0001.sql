BEGIN;
--run the following with superuser privileges:
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS vector;

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
	enable_semantic_search BOOLEAN NOT NULL DEFAULT TRUE,
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


-- videos table start

CREATE TABLE IF NOT EXISTS videos (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	thumb_url TEXT NOT NULL,
	image_url TEXT NOT NULL,
	video_url TEXT NOT NULL,
	subtitles_url TEXT NOT NULL,
	description TEXT NOT NULL,
	release_date DATE NOT NULL,
	width INTEGER NOT NULL,
	height INTEGER NOT NULL,
	duration INTEGER NOT NULL,
	sequence INTEGER NOT NULL DEFAULT 0,
	file TEXT NOT NULL,
	original_file TEXT NOT NULL,
	path TEXT NOT NULL,
	md5sum TEXT NOT NULL,
	enable_semantic_search BOOLEAN NOT NULL DEFAULT TRUE,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- constraint for name length
ALTER TABLE videos
ADD CONSTRAINT videos_check_name_length
CHECK (LENGTH(name) BETWEEN 3 AND 200);

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
CREATE INDEX IF NOT EXISTS movies_title_idx ON videos USING GIN (to_tsvector('simple', name));

-- videos table end

-- actors table start

CREATE TABLE IF NOT EXISTS actors (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	gender TEXT NOT NULL,
	birth_date DATE NOT NULL,
	birth_place TEXT NOT NULL,
	biography TEXT NOT NULL,
	height INTEGER NOT NULL,
	image_url TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- constraint for name length
ALTER TABLE actors
ADD CONSTRAINT actors_check_name_length
CHECK (LENGTH(name) BETWEEN 3 AND 200);

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


-- indexes
CREATE INDEX IF NOT EXISTS songs_title_idx ON actors USING GIN (to_tsvector('simple', name));

ALTER TABLE actors
ADD CONSTRAINT chk_gender_field
CHECK(gender IN ('male', 'female'));

-- actors table end

-- documents table start

CREATE TABLE IF NOT EXISTS documents (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	openai_embeddings vector(1536), -- openai embeddings vector(1536), -- model: text-embedding-ada-002
	st_embeddings vector(512), -- sentence transformer embeddings | model: distiluse-base-multilingual-cased-v1
	google_embeddings vector(768), -- google embeddings | model: embedding-001
	tokens INT NOT NULL DEFAULT 0,
	sequence INT NOT NULL DEFAULT 0,
	content_field TEXT NOT NULL,
	video_id BIGINT NOT NULL REFERENCES videos(id),
	category_id BIGINT NOT NULL REFERENCES categories(id),
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE documents
ADD CONSTRAINT chk_content_field
CHECK (content_field IN ('videos.title', 'videos.description', 'category.name'));

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

INSERT INTO videos (id, name, thumb_url, image_url, video_url, subtitles_url, description, release_date, width, height, duration, sequence, file, original_file, path, md5sum, enable_semantic_search) VALUES (0, 'Unassigned', '', '', '', '', '', '2021-01-01', 0, 0, 0, 0, '', '', '', '', FALSE);
INSERT INTO actors (id, name, gender, birth_date, birth_place, biography, height, image_url) VALUES (0, 'Unassigned', 'male', '2021-01-01', '', '', 0, '');


-- video_categories table start

CREATE TABLE IF NOT EXISTS video_categories (
	video_id BIGINT NOT NULL REFERENCES videos(id),
	category_id BIGINT NOT NULL REFERENCES categories(id),
	PRIMARY KEY (video_id, category_id)
);

-- video_categories table end

-- video_actors table start

CREATE TABLE IF NOT EXISTS video_actors (
	video_id BIGINT NOT NULL REFERENCES videos(id),
	actor_id BIGINT NOT NULL REFERENCES actors(id),
	PRIMARY KEY (video_id, actor_id)
);

-- video_actors table end


-- user table start

CREATE TABLE IF NOT EXISTS users (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	email citext UNIQUE NOT NULL,
	password_hash bytea NOT NULL,
	activated bool NOT NULL,
	version integer NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- auto update modified_at
CREATE OR REPLACE FUNCTION users_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER users_trigger_update_modified_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE users_update_modified_at();

-- user table end

-- tokens table start

CREATE TABLE IF NOT EXISTS tokens (
	hash bytea PRIMARY KEY,
	user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
	expiry timestamp(0) with time zone NOT NULL,
	scope text NOT NULL
);

-- tokens table end

COMMIT;
