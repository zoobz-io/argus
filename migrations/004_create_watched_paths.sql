-- +goose Up
CREATE TABLE watched_paths (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_id BIGINT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_watched_paths_tenant_id ON watched_paths(tenant_id);
CREATE INDEX idx_watched_paths_provider_id ON watched_paths(provider_id);

-- +goose Down
DROP TABLE watched_paths;
