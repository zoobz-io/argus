package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var createHook = rocco.POST[wire.HookCreateRequest, wire.HookCreateResponse]("/hooks", func(r *rocco.Request[wire.HookCreateRequest]) (wire.HookCreateResponse, error) {
	tid := tenantID(r.Identity)
	users := sum.MustUse[contracts.Users](r)
	user, err := users.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return wire.HookCreateResponse{}, ErrUserNotFound
	}

	store := sum.MustUse[contracts.Hooks](r)
	hook, err := store.CreateHook(r, tid, user.ID, r.Body.URL)
	if err != nil {
		return wire.HookCreateResponse{}, err
	}
	return transformers.HookToCreateResponse(hook), nil
}).
	WithSummary("Create a webhook endpoint").
	WithTags("hooks").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(ErrUserNotFound, rocco.ErrValidationFailed)

var listHooks = rocco.GET[rocco.NoBody, wire.HookListResponse]("/hooks", func(r *rocco.Request[rocco.NoBody]) (wire.HookListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Hooks](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListHooksByTenant(r, tid, page)
	if err != nil {
		return wire.HookListResponse{}, err
	}
	return wire.HookListResponse{
		Hooks:  transformers.HooksToResponse(result.Items),
		Offset: result.Offset,
		Limit:  page.PageSize(),
		Total:  result.Total,
	}, nil
}).
	WithSummary("List webhook endpoints").
	WithTags("hooks").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var deleteHook = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/hooks/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Hooks](r)
	if err := store.DeleteHook(r, tid, pathID(r.Params, "id")); err != nil {
		return rocco.NoBody{}, ErrHookNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete a webhook endpoint").
	WithTags("hooks").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrHookNotFound)

var listDeliveries = rocco.GET[rocco.NoBody, wire.DeliveryListResponse]("/hooks/{id}/deliveries", func(r *rocco.Request[rocco.NoBody]) (wire.DeliveryListResponse, error) {
	tid := tenantID(r.Identity)
	hookID := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Deliveries](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListByHook(r, tid, hookID, page)
	if err != nil {
		return wire.DeliveryListResponse{}, ErrHookNotFound
	}
	return transformers.DeliveriesToListResponse(result), nil
}).
	WithSummary("List deliveries for a webhook").
	WithTags("hooks").
	WithPathParams("id").
	WithQueryParams("offset", "limit").
	WithAuthentication().
	WithErrors(ErrHookNotFound)
