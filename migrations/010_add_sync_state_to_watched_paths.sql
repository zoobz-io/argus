-- +goose Up
ALTER TABLE watched_paths ADD COLUMN sync_state TEXT;

-- +goose Down
ALTER TABLE watched_paths DROP COLUMN sync_state;
