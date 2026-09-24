-- Tabel users = akun untuk login. Berbeda dari tabel students
-- (students adalah data yang dilindungi, users adalah yang mengaksesnya).
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL UNIQUE,
    email      VARCHAR(120) NOT NULL,
    password   TEXT         NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'user',
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Pencocokan username saat login tidak membedakan huruf besar/kecil,
-- jadi unique index-nya juga harus begitu.
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_idx
    ON users (LOWER(username));

-- Refresh token disimpan sebagai HASH, bukan nilai aslinya.
-- Alasannya sama seperti password: kalau isi tabel ini bocor,
-- penyerang tetap tidak punya token yang bisa langsung dipakai.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
    ON refresh_tokens (user_id);