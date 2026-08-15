-- 0002_seniors_and_relationships — the care circle.
--
-- A senior profile is the person care is coordinated for. A care relationship
-- connects a user to that senior with a role and a set of permissions, and is
-- the only thing that grants access to their data
-- (docs/02-permissions-and-authorization.md).
--
-- The constraints here reinforce authorization rather than merely describe it:
-- the API cannot store a role, status, or permission the database does not
-- recognise (docs/07-database-and-sync.md).

CREATE TABLE senior_profiles (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The senior's own account, when they have one. NULL for a senior who is
    -- managed by others and never signs in (docs/03-domain-model.md).
    user_id             uuid        UNIQUE REFERENCES users (id) ON DELETE SET NULL,

    -- Who created the profile. Kept distinct from the senior and from circle
    -- membership so coordination can be handed over later without rewriting
    -- history (docs/01-roles-and-care-model.md, "Ownership").
    created_by_user_id  uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,

    display_name        text        NOT NULL,
    date_of_birth       date,
    photo_url           text,
    phone               text,
    address             text,
    emergency_contact   text,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT senior_profiles_display_name_not_blank
        CHECK (length(btrim(display_name)) > 0),
    -- A date of birth in the future is a typo, not a fact.
    CONSTRAINT senior_profiles_date_of_birth_in_past
        CHECK (date_of_birth IS NULL OR date_of_birth <= current_date)
);

CREATE INDEX senior_profiles_created_by_idx ON senior_profiles (created_by_user_id);

CREATE TRIGGER senior_profiles_set_updated_at
    BEFORE UPDATE ON senior_profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE care_relationships (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    senior_id    uuid        NOT NULL REFERENCES senior_profiles (id) ON DELETE CASCADE,
    user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    role         text        NOT NULL,
    -- Permissions are stored per relationship, not derived from the role at
    -- request time, so a circle can narrow or widen one person's access
    -- without inventing new roles.
    permissions  text[]      NOT NULL DEFAULT '{}',
    status       text        NOT NULL DEFAULT 'active',

    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- One relationship per user per senior. A revoked membership is updated
    -- rather than deleted, so care history keeps its author.
    CONSTRAINT care_relationships_unique_member UNIQUE (senior_id, user_id),

    CONSTRAINT care_relationships_role_recognised
        CHECK (role IN ('senior', 'family_member', 'professional_caregiver')),

    CONSTRAINT care_relationships_status_recognised
        CHECK (status IN ('pending', 'active', 'revoked')),

    -- Mirrors internal/care.Permissions. The API cannot persist a permission
    -- the domain does not define, whatever a client sends.
    CONSTRAINT care_relationships_permissions_recognised
        CHECK (permissions <@ ARRAY[
            'senior.view',
            'senior.edit',
            'tasks.view',
            'tasks.manage',
            'tasks.complete',
            'medications.view',
            'medications.manage',
            'medications.record',
            'appointments.view',
            'appointments.manage',
            'notes.view',
            'notes.create',
            'activity.view',
            'members.view',
            'members.invite',
            'members.manage',
            'messages.participate'
        ]::text[])
);

-- "Which seniors can this user reach?" — the professional caregiver's home
-- screen and every authorization check start here.
CREATE INDEX care_relationships_user_status_idx
    ON care_relationships (user_id, status);

-- "Who is in this senior's care circle?"
CREATE INDEX care_relationships_senior_status_idx
    ON care_relationships (senior_id, status);

-- A care circle has exactly one senior. Revoked rows are excluded so the seat
-- can be reassigned if a membership is ever withdrawn.
CREATE UNIQUE INDEX care_relationships_one_senior_per_circle
    ON care_relationships (senior_id)
    WHERE role = 'senior' AND status <> 'revoked';

CREATE TRIGGER care_relationships_set_updated_at
    BEFORE UPDATE ON care_relationships
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
