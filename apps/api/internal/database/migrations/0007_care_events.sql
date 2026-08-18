-- 0007_care_events — one chronological record of what has happened in a
-- senior's care, across every domain.
--
-- Follows docs/03-domain-model.md's CareEvent exactly: id, senior_id,
-- actor_user_id, event_type, entity_type, entity_id, metadata, occurred_at.
--
-- One table, not one per domain. A family member asking "what happened
-- yesterday?" is asking about their relative, not about tasks; four feeds would
-- answer a question nobody has and would have to be merged in the client to
-- answer the one everybody does (plans/phase7.md, objective).
--
-- Events do not replace domain history. The medication history still says what
-- happened to every dose, in medication's own terms; this says what happened in
-- the care, in the circle's terms (plans/phase7.md §8).

CREATE TABLE care_events (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    senior_id           uuid        NOT NULL REFERENCES senior_profiles (id) ON DELETE CASCADE,

    -- The authenticated user who performed the action, never a value from a
    -- request body (plans/phase7.md §4). NULL is reserved for an event no
    -- person performed; nothing fabricates a user to fill it.
    --
    -- ON DELETE SET NULL rather than RESTRICT: the event is the record of what
    -- happened and must outlive the account that did it. Losing the name is a
    -- smaller loss than losing the event.
    actor_user_id       uuid        REFERENCES users (id) ON DELETE SET NULL,

    event_type          text        NOT NULL,

    -- What the event is about, so a client can route to the right screen
    -- without parsing the event type. Deliberately not a foreign key: an event
    -- points at five different tables, and a care event must survive its
    -- subject being deleted — that is what makes it a historical record rather
    -- than a view (plans/phase7.md §5).
    entity_type         text        NOT NULL,
    entity_id           uuid        NOT NULL,

    -- The small amount of structured detail the timeline needs to render a
    -- sentence: a task's title, a medicine's name and dosage. A copy, on
    -- purpose — renaming a task next month must not rewrite what last week's
    -- entry says happened (plans/phase7.md §§5, 9).
    --
    -- jsonb rather than json so the shape is validated on write and the
    -- constraint below can inspect it.
    metadata            jsonb       NOT NULL DEFAULT '{}'::jsonb,

    -- When the thing happened, which is not always when the row was written:
    -- an offline task completion is sent later and the domain mutation decides
    -- the instant. created_at keeps the write time for anybody debugging.
    occurred_at         timestamptz NOT NULL DEFAULT now(),
    created_at          timestamptz NOT NULL DEFAULT now(),

    -- The documented vocabulary, and nothing else. Mirrors
    -- internal/careevents.Types; a name outside it cannot reach the table even
    -- if some future caller tries (plans/phase7.md §25).
    CONSTRAINT care_events_type_recognised
        CHECK (event_type IN (
            'MEMBER_INVITED', 'MEMBER_JOINED', 'MEMBER_REVOKED',
            'TASK_CREATED', 'TASK_COMPLETED', 'TASK_SKIPPED', 'TASK_MISSED',
            'MEDICATION_CREATED', 'MEDICATION_TAKEN', 'MEDICATION_SKIPPED',
            'MEDICATION_MISSED',
            'APPOINTMENT_CREATED', 'APPOINTMENT_COMPLETED', 'APPOINTMENT_CANCELLED',
            'NOTE_ADDED'
        )),

    CONSTRAINT care_events_entity_type_recognised
        CHECK (entity_type IN (
            'task', 'medication', 'appointment', 'relationship', 'invitation', 'note'
        )),

    -- Metadata is a flat object of short labels. Refusing an array or a scalar
    -- at the boundary is what stops it drifting into a copy of the record it
    -- describes (plans/phase7.md §9).
    CONSTRAINT care_events_metadata_is_an_object
        CHECK (jsonb_typeof(metadata) = 'object')
);

-- "What has happened for this senior, newest first?" — the only way the
-- timeline is read.
--
-- id is the third column because the feed pages by keyset on
-- (occurred_at, id): several events can share an instant — a task completed and
-- its event written in the same transaction, or a bulk sync draining a queue —
-- and without a stable tie-break a page boundary would drop or repeat one.
-- A btree is read in either direction, so this one index serves the descending
-- feed (plans/phase7.md §§12, 24).
CREATE INDEX care_events_senior_timeline_idx
    ON care_events (senior_id, occurred_at DESC, id DESC);

-- Deliberately no index on event_type or actor_user_id: nothing filters by
-- either. Filtering the timeline is not a documented requirement, and an index
-- for a query nobody makes is a write cost for nothing (plans/phase7.md §24).

-- No updated_at trigger, and no updated_at column. A care event is written once
-- and never changed; giving it an edit timestamp would suggest otherwise
-- (plans/phase7.md §5).
