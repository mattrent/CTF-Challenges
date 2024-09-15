CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
   id uuid default uuid_generate_v4() PRIMARY KEY,
   username varchar(255) UNIQUE NOT NULL,
   password_hash varchar(255) NOT NULL,
   role varchar(255) NULL,
   created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS challenges (
   id uuid default uuid_generate_v4() PRIMARY KEY,
   user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   published boolean NOT NULL default false
);

CREATE TABLE IF NOT EXISTS instances (
   id uuid default uuid_generate_v4() PRIMARY KEY,
   challenge_id uuid NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
   user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   token varchar(255) NOT NULL,
   created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tests (
   id uuid default uuid_generate_v4() PRIMARY KEY,
   challenge_id uuid NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
   status_code int NOT NULL,
   log varchar NOT NULL,
   created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);
