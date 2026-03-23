-- +goose Up
CREATE TABLE documents (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_id BIGINT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    watched_path_id BIGINT NOT NULL REFERENCES watched_paths(id) ON DELETE CASCADE,
    current_version_id BIGINT,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_documents_provider_id ON documents(provider_id);
CREATE INDEX idx_documents_watched_path_id ON documents(watched_path_id);
CREATE UNIQUE INDEX idx_documents_external_id ON documents(tenant_id, provider_id, external_id);

-- +goose Down
DROP TABLE documents;
