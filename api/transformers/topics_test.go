//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestTopicToResponse(t *testing.T) {
	topic := argustest.NewTopic()
	resp := TopicToResponse(topic)

	if resp.ID != "tp1" || resp.Name != "Security" || resp.Description != "Sec docs" {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestTopicsToListResponse(t *testing.T) {
	topics := []*models.Topic{argustest.NewTopic(), argustest.NewTopic()}
	topics[1].ID = "tp2"

	resp := TopicsToListResponse(topics)
	if len(resp.Topics) != 2 || resp.Topics[1].ID != "tp2" {
		t.Errorf("unexpected result: %+v", resp)
	}
}

func TestTopicsToListResponse_Empty(t *testing.T) {
	resp := TopicsToListResponse(nil)
	if len(resp.Topics) != 0 {
		t.Errorf("expected empty, got %d", len(resp.Topics))
	}
}
