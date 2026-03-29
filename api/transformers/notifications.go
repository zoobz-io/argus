package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// NotificationToResponse transforms a Notification model to an API response.
func NotificationToResponse(n *models.Notification) wire.NotificationResponse {
	return wire.NotificationResponse{
		ID:         n.ID,
		Type:       n.Type,
		Status:     n.Status,
		CreatedAt:  n.CreatedAt,
		DocumentID: n.DocumentID,
		VersionID:  n.VersionID,
		Message:    n.Message,
	}
}

// NotificationsToListResponse transforms an OffsetResult of notifications to a list response.
func NotificationsToListResponse(result *models.OffsetResult[models.Notification]) wire.NotificationListResponse {
	notifications := make([]wire.NotificationResponse, len(result.Items))
	for i, n := range result.Items {
		notifications[i] = NotificationToResponse(n)
	}
	return wire.NotificationListResponse{
		Notifications: notifications,
		Offset:        result.Offset,
		Limit:         len(result.Items),
		Total:         result.Total,
	}
}
