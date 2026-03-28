//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWatchedPathsToListResponse(t *testing.T) {
	result := &models.OffsetResult[models.WatchedPath]{
		Items:  []*models.WatchedPath{argustest.NewWatchedPath()},
		Offset: 0,
		Total:  1,
	}

	resp := WatchedPathsToListResponse(result)
	if len(resp.WatchedPaths) != 1 || resp.WatchedPaths[0].ID != "wp1" {
		t.Errorf("paths mismatch: %+v", resp.WatchedPaths)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: %+v", resp)
	}
}
