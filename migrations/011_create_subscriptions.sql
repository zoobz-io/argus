-- +goose Up
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    channel TEXT NOT NULL DEFAULT 'inbox',
    filters JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, event_type, channel)
);

CREATE INDEX idx_subscriptions_tenant_event ON subscriptions(tenant_id, event_type);
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);

-- +goose Down
DROP TABLE subscriptions;
