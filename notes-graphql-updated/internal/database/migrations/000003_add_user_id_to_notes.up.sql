ALTER TABLE notes
    ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- In a fresh dev database notes is empty, so NOT NULL can be enforced
-- immediately. If you already have note rows, backfill user_id for each
-- one first, then run this ALTER separately.
ALTER TABLE notes ALTER COLUMN user_id SET NOT NULL;

CREATE INDEX idx_notes_user_id ON notes (user_id);
