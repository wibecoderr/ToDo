-- USERS AUDIT COLUMNS
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- TODO TABLE
DROP TABLE IF EXISTS todo;

CREATE TABLE todo (
                      id SERIAL PRIMARY KEY,
                      user_id UUID NOT NULL REFERENCES users(uid) ,
                      title VARCHAR(255) NOT NULL,
                      description TEXT NOT NULL,
                      status status_enum NOT NULL DEFAULT 'active',
                      created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                      updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                      archived_at TIMESTAMPTZ,
                      deadline TIMESTAMPTZ
);

-- USER SESSION TABLE
DROP TABLE IF EXISTS user_session;

CREATE TABLE user_session (
                              id SERIAL PRIMARY KEY,
                              user_id UUID NOT NULL REFERENCES users(uid) ,
                              session_token TEXT NOT NULL UNIQUE,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                              expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '24 hours')
);
