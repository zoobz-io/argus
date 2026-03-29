package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listMySubscriptions = rocco.GET[rocco.NoBody, wire.SubscriptionListResponse]("/subscriptions", func(r *rocco.Request[rocco.NoBody]) (wire.SubscriptionListResponse, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return wire.SubscriptionListResponse{}, ErrUserNotFound
	}

	store := sum.MustUse[contracts.Subscriptions](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListSubscriptionsByUser(r, tid, user.ID, page)
	if err != nil {
		return wire.SubscriptionListResponse{}, err
	}
	return wire.SubscriptionListResponse{
		Subscriptions: transformers.SubscriptionsToResponse(result.Items),
		Offset:        result.Offset,
		Limit:         page.PageSize(),
		Total:         result.Total,
	}, nil
}).
	WithSummary("List my subscriptions").
	WithTags("subscriptions").
	WithQueryParams("offset", "limit").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var createSubscription = rocco.POST[wire.SubscriptionRequest, wire.SubscriptionResponse]("/subscriptions", func(r *rocco.Request[wire.SubscriptionRequest]) (wire.SubscriptionResponse, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return wire.SubscriptionResponse{}, ErrUserNotFound
	}

	store := sum.MustUse[contracts.Subscriptions](r)
	sub, err := store.CreateSubscription(r, tid, user.ID, r.Body.EventType, r.Body.Channel, r.Body.WebhookEndpointID)
	if err != nil {
		return wire.SubscriptionResponse{}, err
	}
	return transformers.SubscriptionToResponse(sub), nil
}).
	WithSummary("Create a subscription").
	WithTags("subscriptions").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var deleteSubscription = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/subscriptions/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return rocco.NoBody{}, ErrUserNotFound
	}
	store := sum.MustUse[contracts.Subscriptions](r)
	if err := store.DeleteSubscription(r, tid, user.ID, pathID(r.Params, "id")); err != nil {
		return rocco.NoBody{}, ErrSubscriptionNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete a subscription").
	WithTags("subscriptions").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrSubscriptionNotFound)
