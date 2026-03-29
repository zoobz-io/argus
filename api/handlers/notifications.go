package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listNotifications = rocco.GET[rocco.NoBody, wire.NotificationListResponse]("/notifications", func(r *rocco.Request[rocco.NoBody]) (wire.NotificationListResponse, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return wire.NotificationListResponse{}, ErrUserNotFound
	}

	store := sum.MustUse[contracts.Notifications](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.SearchByUser(r, tid, user.ID, page)
	if err != nil {
		return wire.NotificationListResponse{}, err
	}
	return transformers.NotificationsToListResponse(result), nil
}).
	WithSummary("List my notifications").
	WithTags("notifications").
	WithQueryParams("offset", "limit", "status").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var updateNotificationStatus = rocco.PATCH[wire.NotificationUpdateRequest, wire.NotificationResponse]("/notifications/{id}", func(r *rocco.Request[wire.NotificationUpdateRequest]) (wire.NotificationResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Notifications](r)
	n, err := store.UpdateStatus(r, tid, pathID(r.Params, "id"), models.NotificationStatus(r.Body.Status))
	if err != nil {
		return wire.NotificationResponse{}, ErrNotificationNotFound
	}
	return transformers.NotificationToResponse(n), nil
}).
	WithSummary("Update notification status").
	WithTags("notifications").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrNotificationNotFound)

var bulkMarkRead = rocco.POST[wire.NotificationBulkReadRequest, rocco.NoBody]("/notifications/mark-read", func(r *rocco.Request[wire.NotificationBulkReadRequest]) (rocco.NoBody, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return rocco.NoBody{}, ErrUserNotFound
	}

	store := sum.MustUse[contracts.Notifications](r)
	if err := store.BulkUpdateStatus(r, tid, user.ID, models.NotificationStatus(r.Body.Status)); err != nil {
		return rocco.NoBody{}, err
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Bulk update notification status").
	WithTags("notifications").
	WithAuthentication().
	WithErrors(ErrUserNotFound)
