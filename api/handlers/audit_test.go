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

func TestListAuditLog_Success(t *testing.T) {
	ts := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	mock := &argustest.MockAuditLog{
		OnSearch: func(_ context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			if params.TenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", params.TenantID)
			}
			return &models.OffsetResult[models.DomainEvent]{
				Items: []*models.DomainEvent{{
					ID:           "a-1",
					Timestamp:    ts,
					Action:       "provider.created",
					ResourceType: "provider",
					ResourceID:   "p-1",
					ActorID:      "u-1",
					TenantID:     "tenant-1",
					Metadata:     json.RawMessage(`{"provider_type":"google_drive"}`),
				}},
				Total: 1,
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIAuditLog(mock),
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
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
}

func TestListAuditLog_WithQueryParams(t *testing.T) {
	mock := &argustest.MockAuditLog{
		OnSearch: func(_ context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			if params.Action != "provider.created" {
				t.Errorf("Action = %q, want provider.created", params.Action)
			}
			if params.ResourceType != "provider" {
				t.Errorf("ResourceType = %q, want provider", params.ResourceType)
			}
			if params.ActorID != "u-1" {
				t.Errorf("ActorID = %q, want u-1", params.ActorID)
			}
			if params.Offset != 5 {
				t.Errorf("Offset = %d, want 5", params.Offset)
			}
			if params.Limit != 10 {
				t.Errorf("Limit = %d, want 10", params.Limit)
			}
			return &models.OffsetResult[models.DomainEvent]{Items: []*models.DomainEvent{}, Total: 0}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit?action=provider.created&resource_type=provider&actor_id=u-1&offset=5&limit=10", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestListAuditLog_Error(t *testing.T) {
	mock := &argustest.MockAuditLog{
		OnSearch: func(_ context.Context, _ models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error) {
			return nil, fmt.Errorf("search error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIAuditLog(mock),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/audit", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}
