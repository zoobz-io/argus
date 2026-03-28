//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestTopicToAdminResponse(t *testing.T) {
	topic := argustest.NewTopic()
	resp := TopicToAdminResponse(topic)

	if resp.ID != "tp1" || resp.TenantID != "t1" || resp.Name != "Security" || resp.Description != "Sec docs" {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestTopicsToAdminList(t *testing.T) {
	topics := []*models.Topic{argustest.NewTopic(), argustest.NewTopic()}
	topics[1].ID = "tp2"

	resp := TopicsToAdminList(topics)
	if len(resp.Topics) != 2 || resp.Topics[1].ID != "tp2" {
		t.Errorf("topics mismatch: %+v", resp.Topics)
	}
}
