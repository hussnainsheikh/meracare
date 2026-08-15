-- 0001_init — foundation schema.
--
-- Establishes the application `users` table and the shared `updated_at`
-- trigger. Supabase's `auth.users` remains separate: the application links to
-- it by `auth_user_id` only (docs/07-database-and-sync.md).
--
-- No foreign key is declared against `auth.users` so that local Postgres and
-- CI databases, which have no Supabase auth schema, run the same migrations.

-- gen_random_uuid() is built into PostgreSQL 13+; pgcrypto is requested only as
-- a safety net for older local images.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Keeps updated_at honest without relying on every writer to set it.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Supabase auth.users.id. One application user per authenticated identity.
    auth_user_id  uuid        NOT NULL UNIQUE,
    email         text,
    display_name  text        NOT NULL,
    avatar_url    text,
    phone         text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT users_display_name_not_blank CHECK (length(btrim(display_name)) > 0),
    CONSTRAINT users_email_not_blank CHECK (email IS NULL OR length(btrim(email)) > 0)
);

-- Invitations are matched by email, so lookups must be case-insensitive and
-- an address must not map to two application users.
CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email)) WHERE email IS NOT NULL;

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
