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
		OnSearch: func(_ context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
			return &models.OffsetResult[models.AuditEntry]{
				Items: []*models.AuditEntry{{
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
		OnSearch: func(_ context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
			if params.TenantID != "t-1" {
				t.Errorf("TenantID = %q, want t-1", params.TenantID)
			}
			if params.Action != "provider.created" {
				t.Errorf("Action = %q, want provider.created", params.Action)
			}
			return &models.OffsetResult[models.AuditEntry]{Items: []*models.AuditEntry{}, Total: 0}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(),
		argustest.WithAdminAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit?tenant_id=t-1&action=provider.created", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestAdminListAuditLog_Error(t *testing.T) {
	mock := &argustest.MockAdminAuditLog{
		OnSearch: func(_ context.Context, _ models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
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
