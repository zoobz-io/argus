-- +goose Up
CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    content_hash TEXT NOT NULL,
    extraction_status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX idx_document_versions_tenant_id ON document_versions(tenant_id);
CREATE UNIQUE INDEX idx_document_versions_document_version ON document_versions(document_id, version_number);
CREATE INDEX idx_document_versions_content_hash ON document_versions(content_hash);

ALTER TABLE documents ADD CONSTRAINT fk_documents_current_version
    FOREIGN KEY (current_version_id) REFERENCES document_versions(id);

-- +goose Down
ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_documents_current_version;
DROP TABLE document_versions;
