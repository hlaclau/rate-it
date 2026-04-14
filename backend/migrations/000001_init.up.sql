CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE media_source AS ENUM ('tmdb', 'imdb', 'openlibrary');
CREATE TYPE media_type   AS ENUM ('movie', 'series', 'book');
CREATE TYPE media_status AS ENUM ('watched', 'plan_to_watch');

-- Users
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        NOT NULL UNIQUE,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    bio           TEXT,
    avatar_url    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Media: minimal local record for FK integrity and card rendering.
-- Full details come from the external source (e.g. TMDB) and are cached in Redis.
CREATE TABLE media (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id  TEXT         NOT NULL,
    source       media_source NOT NULL,
    type         media_type   NOT NULL,
    title        TEXT         NOT NULL,
    poster_path  TEXT,
    release_year SMALLINT,
    cached_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (external_id, source)
);

-- One row per user per media item: tracks status and optional rating.
CREATE TABLE user_media (
    user_id   UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_id  UUID         NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    status    media_status NOT NULL,
    rating    SMALLINT     CHECK (rating BETWEEN 1 AND 10), -- nullable; only set when watched
    added_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, media_id)
);

CREATE INDEX ON media (source, type);
CREATE INDEX ON user_media (user_id);
CREATE INDEX ON user_media (media_id);
