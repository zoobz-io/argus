-- +goose Up
CREATE TABLE hooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hooks_tenant_id ON hooks(tenant_id);
CREATE INDEX idx_hooks_user_id ON hooks(user_id);

CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hook_id UUID NOT NULL REFERENCES hooks(id) ON DELETE CASCADE,
    event_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    status_code INTEGER NOT NULL,
    error TEXT,
    attempt INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_deliveries_hook_id ON deliveries(hook_id);
CREATE INDEX idx_deliveries_tenant_id ON deliveries(tenant_id);

ALTER TABLE subscriptions ADD COLUMN webhook_endpoint_id UUID REFERENCES hooks(id) ON DELETE SET NULL;
CREATE INDEX idx_subscriptions_webhook_endpoint_id ON subscriptions(webhook_endpoint_id);

-- +goose Down
ALTER TABLE subscriptions DROP COLUMN webhook_endpoint_id;
DROP TABLE deliveries;
DROP TABLE hooks;
