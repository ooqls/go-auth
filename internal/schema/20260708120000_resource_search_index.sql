-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_resourcesv1_name_trgm ON resourcesv1 USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_resourcesv1_group_kind ON resourcesv1 (rGroup, kind);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_resourcesv1_group_kind;
DROP INDEX IF EXISTS idx_resourcesv1_name_trgm;

-- +goose StatementEnd
