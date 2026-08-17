-- 0004_care_tasks — the daily care routine.
--
-- Two tables, following docs/03-domain-model.md exactly:
--
--   care_task_templates  the rule ("every weekday at 09:00")
--   care_task_instances  one concrete occurrence that can be completed or skipped
--
-- A one-time task is an instance with no template. That keeps a single table as
-- the answer to "what is due today", whether the task recurs or not.
--
-- Instances carry their own title and description rather than reading through
-- to the template. A completed task must keep the wording it had when it was
-- completed: renaming "Morning walk" to "Afternoon walk" must not rewrite what
-- somebody already did (plans/phase4.md §18).

-- The senior's timezone decides when their day starts and ends, so "today's
-- tasks" and every recurrence are computed in it. Without this the server would
-- have to guess, and a task due at 09:00 would land at the wrong hour for
-- anybody outside UTC (plans/phase4.md §14).
ALTER TABLE senior_profiles
    ADD COLUMN timezone text NOT NULL DEFAULT 'UTC';

ALTER TABLE senior_profiles
    ADD CONSTRAINT senior_profiles_timezone_not_blank
        CHECK (length(btrim(timezone)) > 0);

CREATE TABLE care_task_templates (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    senior_id           uuid        NOT NULL REFERENCES senior_profiles (id) ON DELETE CASCADE,

    -- Kept for the audit trail, so the row survives the creator leaving the
    -- circle, exactly as invitations.inviter_user_id does.
    created_by_user_id  uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,

    title               text        NOT NULL,
    description         text        NOT NULL DEFAULT '',

    -- NULL means nobody in particular: whoever is around does it.
    assigned_user_id    uuid        REFERENCES users (id) ON DELETE SET NULL,

    -- An RRULE subset: 'FREQ=DAILY' or 'FREQ=WEEKLY;BYDAY=MO,WE,FR'.
    -- internal/tasks.ParseRecurrence is the authority on what parses.
    recurrence_rule     text        NOT NULL,

    -- Wall-clock time in the senior's timezone, not an instant. 09:00 means
    -- "nine in the morning where they live", which is the only reading that
    -- survives a daylight-saving change.
    due_time            time        NOT NULL,

    -- Deactivated rather than deleted: the instances it already produced are
    -- care history (plans/phase4.md §19).
    active              boolean     NOT NULL DEFAULT true,

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT care_task_templates_title_not_blank
        CHECK (length(btrim(title)) > 0),

    -- Lets instances point at (template, senior) as a pair, below.
    CONSTRAINT care_task_templates_id_senior_key UNIQUE (id, senior_id)
);

-- "Which routines does this senior have?" — the only way templates are listed.
CREATE INDEX care_task_templates_senior_idx
    ON care_task_templates (senior_id)
    WHERE active;

CREATE TRIGGER care_task_templates_set_updated_at
    BEFORE UPDATE ON care_task_templates
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE care_task_instances (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- NULL for a one-time task. ON DELETE CASCADE never fires in practice:
    -- templates are deactivated, not deleted.
    template_id         uuid        REFERENCES care_task_templates (id) ON DELETE CASCADE,
    senior_id           uuid        NOT NULL REFERENCES senior_profiles (id) ON DELETE CASCADE,

    created_by_user_id  uuid        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,

    title               text        NOT NULL,
    description         text        NOT NULL DEFAULT '',

    assigned_user_id    uuid        REFERENCES users (id) ON DELETE SET NULL,

    -- The absolute instant the task is due, derived from the senior's local
    -- date and time at the moment the occurrence was worked out.
    scheduled_for       timestamptz NOT NULL,

    status              text        NOT NULL DEFAULT 'pending',

    completed_at        timestamptz,
    completed_by        uuid        REFERENCES users (id) ON DELETE SET NULL,

    -- docs/03 names completed_at/completed_by but nothing for a skip, and
    -- plans/phase4.md §11 requires both who and when. A skip is a deliberate
    -- care decision, so it is attributed as carefully as a completion.
    skipped_at          timestamptz,
    skipped_by          uuid        REFERENCES users (id) ON DELETE SET NULL,

    -- Free text the actor may add when completing or skipping.
    notes               text        NOT NULL DEFAULT '',

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT care_task_instances_title_not_blank
        CHECK (length(btrim(title)) > 0),

    -- 'overdue' is deliberately absent. It is a reading of the clock against
    -- scheduled_for, not a state anybody moves a task into, so storing it would
    -- mean a background job had to run for the data to be true
    -- (plans/phase4.md §12).
    CONSTRAINT care_task_instances_status_recognised
        CHECK (status IN ('pending', 'completed', 'skipped', 'cancelled')),

    -- A completion or a skip must say who and when, so the audit trail cannot
    -- be half-written. Same shape as invitations_accepted_is_attributed.
    CONSTRAINT care_task_instances_completed_is_attributed
        CHECK (
            status <> 'completed'
            OR (completed_at IS NOT NULL AND completed_by IS NOT NULL)
        ),

    CONSTRAINT care_task_instances_skipped_is_attributed
        CHECK (
            status <> 'skipped'
            OR (skipped_at IS NOT NULL AND skipped_by IS NOT NULL)
        ),

    -- An occurrence exists at most once per template. This is what makes
    -- materialising a window idempotent: two concurrent readers of the same
    -- date range race, and the loser's insert is discarded rather than
    -- producing a duplicate task.
    CONSTRAINT care_task_instances_one_per_occurrence
        UNIQUE (template_id, scheduled_for),

    -- An instance cannot borrow a template from another senior's circle. The
    -- application checks this too; the database makes it impossible.
    CONSTRAINT care_task_instances_template_same_senior
        FOREIGN KEY (template_id, senior_id)
        REFERENCES care_task_templates (id, senior_id)
        ON DELETE CASCADE
);

-- "What is due for this senior between these two instants?" — the query behind
-- today's tasks, upcoming tasks, and the overdue list.
CREATE INDEX care_task_instances_senior_schedule_idx
    ON care_task_instances (senior_id, scheduled_for);

-- "What is assigned to me?" — the home screen, across every circle the user
-- belongs to. Restricted to pending because that screen never asks for
-- finished work.
CREATE INDEX care_task_instances_assignee_idx
    ON care_task_instances (assigned_user_id, scheduled_for)
    WHERE status = 'pending';

-- Supports regenerating a template's future occurrences after an edit.
CREATE INDEX care_task_instances_template_schedule_idx
    ON care_task_instances (template_id, scheduled_for)
    WHERE template_id IS NOT NULL;

CREATE TRIGGER care_task_instances_set_updated_at
    BEFORE UPDATE ON care_task_instances
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
