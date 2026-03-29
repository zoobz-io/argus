//go:build testing

package notify

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

type mockNotifyIndexer struct {
	OnIndex func(ctx context.Context, n *models.Notification) error
}

func (m *mockNotifyIndexer) Index(ctx context.Context, n *models.Notification) error {
	return m.OnIndex(ctx, n)
}

func setupIndexRegistry(t *testing.T, indexer intcontracts.NotifyIndexer) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.NotifyIndexer](k, indexer)
	sum.Freeze(k)
	return context.Background()
}

func TestIndexStage_HappyPath(t *testing.T) {
	var indexed bool
	indexer := &mockNotifyIndexer{
		OnIndex: func(_ context.Context, n *models.Notification) error {
			indexed = true
			if n.ID != "n-1" {
				t.Errorf("notification ID: got %q, want %q", n.ID, "n-1")
			}
			return nil
		},
	}

	ctx := setupIndexRegistry(t, indexer)
	stage := newIndexStage()

	item := &FanOutItem{
		Notification: &models.Notification{
			ID:       "n-1",
			TenantID: "t-1",
			UserID:   "u-1",
		},
	}

	_, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !indexed {
		t.Error("expected indexer.Index to be called")
	}
}

func TestIndexStage_Error(t *testing.T) {
	indexer := &mockNotifyIndexer{
		OnIndex: func(_ context.Context, _ *models.Notification) error {
			return errors.New("index unavailable")
		},
	}

	ctx := setupIndexRegistry(t, indexer)
	stage := newIndexStage()

	item := &FanOutItem{
		Notification: &models.Notification{
			ID: "n-1",
		},
	}

	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Fatal("expected error")
	}
}
