//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestTagToResponse(t *testing.T) {
	tag := argustest.NewTag()
	resp := TagToResponse(tag)

	if resp.ID != "tg1" || resp.Name != "compliance" || resp.Description != "Compliance docs" {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestTagsToListResponse(t *testing.T) {
	tags := []*models.Tag{argustest.NewTag(), argustest.NewTag()}
	tags[1].ID = "tg2"

	resp := TagsToListResponse(tags)
	if len(resp.Tags) != 2 || resp.Tags[1].ID != "tg2" {
		t.Errorf("unexpected result: %+v", resp)
	}
}
