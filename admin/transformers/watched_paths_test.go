//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWatchedPathToAdminResponse(t *testing.T) {
	w := argustest.NewWatchedPath()
	resp := WatchedPathToAdminResponse(w)

	if resp.ID != "wp1" || resp.TenantID != "t1" || resp.Path != "/docs" || !resp.Active {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestWatchedPathsToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.WatchedPath]{
		Items:  []*models.WatchedPath{argustest.NewWatchedPath()},
		Offset: 0,
		Total:  1,
	}

	resp := WatchedPathsToAdminList(result)
	if len(resp.WatchedPaths) != 1 || resp.WatchedPaths[0].ID != "wp1" {
		t.Errorf("paths mismatch: %+v", resp.WatchedPaths)
	}
	if resp.Total != 1 {
		t.Errorf("expected Total=1, got %d", resp.Total)
	}
}
