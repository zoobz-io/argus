//go:build testing

package wire

import (
	"testing"
)

func TestTopicListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := TopicListResponse{
		Topics: []TopicResponse{{ID: "t1", Name: "Security"}},
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTopicListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := TopicListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestTopicListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := TopicListResponse{
		Topics: []TopicResponse{{ID: "t1"}},
	}
	cloned := orig.Clone()

	cloned.Topics[0].ID = "mutated"
	if orig.Topics[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestTopicListResponse_Clone_NilSlice(t *testing.T) {
	orig := TopicListResponse{}
	cloned := orig.Clone()
	if cloned.Topics != nil {
		t.Error("expected nil Topics")
	}
}
