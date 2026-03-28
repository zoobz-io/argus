//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWatchedPathToResponse(t *testing.T) {
	w := argustest.NewWatchedPath()
	resp := WatchedPathToResponse(w)

	if resp.ID != "wp1" || resp.Path != "/docs" || resp.ProviderID != "p1" || !resp.Active {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestWatchedPathsToResponse(t *testing.T) {
	paths := []*models.WatchedPath{argustest.NewWatchedPath(), argustest.NewWatchedPath()}
	paths[1].ID = "wp2"

	result := WatchedPathsToResponse(paths)
	if len(result) != 2 || result[1].ID != "wp2" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestApplyWatchedPathCreate(t *testing.T) {
	req := wire.WatchedPathCreateRequest{Path: "/new/path", ProviderID: "p2"}
	w := &models.WatchedPath{}
	ApplyWatchedPathCreate(req, w)

	if w.Path != "/new/path" || w.ProviderID != "p2" {
		t.Errorf("fields not applied: %+v", w)
	}
}
