//go:build testing

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestAdminListAuditLog_Success(t *testing.T) {
	ts := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	mock := &argustest.MockAdminAuditLog{
		OnSearchAll: func(_ context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			return &models.OffsetResult[models.DomainEvent]{
				Items: []*models.DomainEvent{{
					ID:           "a-1",
					Timestamp:    ts,
					Action:       "tenant.created",
					ResourceType: "tenant",
					ResourceID:   "t-1",
					ActorID:      "admin-1",
					TenantID:     "t-1",
					Metadata:     json.RawMessage(`{"name":"Acme","slug":"acme"}`),
				}},
				Total: 1,
			}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(),
		argustest.WithAdminAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.AuditListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Entries) != 1 || resp.Entries[0].ID != "a-1" {
		t.Errorf("unexpected entries: %+v", resp.Entries)
	}
}

func TestAdminListAuditLog_WithTenantFilter(t *testing.T) {
	mock := &argustest.MockAdminAuditLog{
		OnSearchAll: func(_ context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			if params.TenantID != "t-1" {
				t.Errorf("TenantID = %q, want t-1", params.TenantID)
			}
			if params.Action != "provider.created" {
				t.Errorf("Action = %q, want provider.created", params.Action)
			}
			return &models.OffsetResult[models.DomainEvent]{Items: []*models.DomainEvent{}, Total: 0}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(),
		argustest.WithAdminAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit?tenant_id=t-1&action=provider.created", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestAdminListAuditLog_AllFilters(t *testing.T) {
	mock := &argustest.MockAdminAuditLog{
		OnSearchAll: func(_ context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			if params.ResourceType != "provider" {
				t.Errorf("ResourceType = %q, want provider", params.ResourceType)
			}
			if params.ActorID != "u-1" {
				t.Errorf("ActorID = %q, want u-1", params.ActorID)
			}
			if params.From == nil {
				t.Error("From should be set")
			}
			if params.To == nil {
				t.Error("To should be set")
			}
			if params.Limit != 50 {
				t.Errorf("Limit = %d, want 50", params.Limit)
			}
			if params.Offset != 10 {
				t.Errorf("Offset = %d, want 10", params.Offset)
			}
			return &models.OffsetResult[models.DomainEvent]{Items: []*models.DomainEvent{}, Total: 0}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(),
		argustest.WithAdminAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit?resource_type=provider&actor_id=u-1&from=2026-03-01T00:00:00Z&to=2026-03-31T00:00:00Z&limit=50&offset=10", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestAdminListAuditLog_Error(t *testing.T) {
	mock := &argustest.MockAdminAuditLog{
		OnSearchAll: func(_ context.Context, _ models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			return nil, fmt.Errorf("search error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(),
		argustest.WithAdminAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}
