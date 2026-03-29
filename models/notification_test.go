package models

import (
	"testing"
	"time"
)

func TestNotification_Clone(t *testing.T) {
	n := Notification{
		ID:         "n-1",
		UserID:     "u-1",
		EventID:    "evt-1",
		TenantID:   "t-1",
		DocumentID: "doc-1",
		Type:       NotificationIngestCompleted,
		Status:     NotificationUnread,
		Message:    "Document ingestion completed",
		CreatedAt:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	clone := n.Clone()

	if clone.ID != n.ID || clone.UserID != n.UserID || clone.Type != n.Type || clone.Message != n.Message {
		t.Error("Clone did not copy all fields")
	}

	clone.Message = "mutated"
	if n.Message != "Document ingestion completed" {
		t.Error("mutating clone leaked to original")
	}
}
