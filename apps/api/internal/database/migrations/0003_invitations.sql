-- 0003_invitations — inviting people into a senior's care circle.
--
-- An invitation is a proposal: a role and a permission set that become a real
-- care relationship only when the recipient accepts (docs/04, "Invitation
-- Workflow"). Nothing here grants access on its own.
--
-- The raw invitation token is never stored. Only its SHA-256 hash is kept, so
-- a database disclosure does not hand an attacker working invitations
-- (docs/09-security-privacy.md).

CREATE TABLE invitations (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    senior_id        uuid        NOT NULL REFERENCES senior_profiles (id) ON DELETE CASCADE,

    -- The inviter is kept for the audit trail docs/02 requires, so the row
    -- survives even if that person later leaves the circle.
    inviter_user_id  uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,

    -- MVP invites by email. docs/03 allows phone as an alternative; adding it
    -- later means a nullable column and a relaxed CHECK, not a new table.
    invitee_email    text        NOT NULL,

    role             text        NOT NULL,
    -- The permissions the recipient will hold once they accept. Validated
    -- against the inviter's own permissions before insert.
    permissions      text[]      NOT NULL DEFAULT '{}',

    -- SHA-256 of the raw token. The token carries 256 bits of entropy, so a
    -- plain cryptographic hash is the right primitive here — password hashes
    -- like bcrypt exist to slow brute force against low-entropy secrets, which
    -- this is not.
    token_hash       bytea       NOT NULL,

    status           text        NOT NULL DEFAULT 'pending',
    expires_at       timestamptz NOT NULL,

    accepted_at         timestamptz,
    accepted_by_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
    revoked_at          timestamptz,

    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT invitations_token_hash_key UNIQUE (token_hash),

    CONSTRAINT invitations_email_not_blank
        CHECK (length(btrim(invitee_email)) > 0),

    -- A circle has exactly one senior and that seat is not invitable; only
    -- family members and professional caregivers can be invited in.
    CONSTRAINT invitations_role_recognised
        CHECK (role IN ('family_member', 'professional_caregiver')),

    CONSTRAINT invitations_status_recognised
        CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),

    -- Mirrors internal/care.Permissions, exactly as care_relationships does.
    CONSTRAINT invitations_permissions_recognised
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
        ]::text[]),

    -- An accepted invitation must record who accepted it and when, so the
    -- audit trail cannot be half-written.
    CONSTRAINT invitations_accepted_is_attributed
        CHECK (
            status <> 'accepted'
            OR (accepted_at IS NOT NULL AND accepted_by_user_id IS NOT NULL)
        ),

    CONSTRAINT invitations_revoked_is_timestamped
        CHECK (status <> 'revoked' OR revoked_at IS NOT NULL)
);

-- Token lookup is the hot path on accept, and must be a single indexed probe.
-- (The UNIQUE constraint above already provides the index.)

-- "Which invitations are outstanding for this senior?"
CREATE INDEX invitations_senior_status_idx ON invitations (senior_id, status);

-- "Has this person already been invited?" — matched case-insensitively,
-- because email addresses are not case-sensitive in practice.
CREATE INDEX invitations_email_idx ON invitations (lower(invitee_email), status);

CREATE INDEX invitations_inviter_idx ON invitations (inviter_user_id);

-- Supports sweeping expired invitations, and keeps the expiry check cheap.
CREATE INDEX invitations_expires_at_idx ON invitations (expires_at)
    WHERE status = 'pending';

-- One outstanding invitation per person per circle. Re-inviting someone
-- revokes or reuses the existing invitation rather than accumulating duplicates
-- that would each grant access.
CREATE UNIQUE INDEX invitations_one_pending_per_invitee
    ON invitations (senior_id, lower(invitee_email))
    WHERE status = 'pending';

CREATE TRIGGER invitations_set_updated_at
    BEFORE UPDATE ON invitations
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
