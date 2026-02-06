BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ENUM
CREATE TYPE status_enum AS ENUM (
    'active',
   'completed',
    'incomplete'
    );

-- USERS TABLE
CREATE TABLE users (
                       uid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       name VARCHAR(100) NOT NULL,
                       email VARCHAR(100) UNIQUE NOT NULL,
                       phone_number VARCHAR(10) UNIQUE NOT NULL,
                       password TEXT NOT NULL,
                       age INT NOT NULL CHECK (age BETWEEN 1 AND 120)



);


CREATE TABLE todo (
                      id SERIAL PRIMARY KEY,
                      user_id UUID NOT NULL REFERENCES users(uid) ON DELETE CASCADE,
                      title VARCHAR(20) NOT NULL,
                      description VARCHAR(100) NOT NULL,
                      created_at TIMESTAMP DEFAULT now(),
                      deadline TIMESTAMPTZ NOT NULL,
                      status status_enum NOT NULL DEFAULT 'active'
);

-- USER SESSION TABLE
CREATE TABLE user_session (
                              id SERIAL PRIMARY KEY,
                              user_id UUID NOT NULL REFERENCES users(uid),
                              session_token TEXT NOT NULL,
                              created_at TIMESTAMPTZ DEFAULT now()
);

COMMIT;
