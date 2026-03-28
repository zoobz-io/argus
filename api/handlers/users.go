package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listUsersInTenant = rocco.GET[rocco.NoBody, wire.UserListResponse]("/users", func(r *rocco.Request[rocco.NoBody]) (wire.UserListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Users](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListUsersByTenant(r, tid, page)
	if err != nil {
		return wire.UserListResponse{}, err
	}
	return wire.UserListResponse{
		Users:  transformers.UsersToResponse(result.Items),
		Offset: result.Offset,
		Total:  result.Total,
		Limit:  page.PageSize(),
	}, nil
}).
	WithSummary("List users in tenant").
	WithTags("users").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getMyProfile = rocco.GET[rocco.NoBody, wire.UserResponse]("/users/me", func(r *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
	store := sum.MustUse[contracts.Users](r)
	user, err := store.GetUserByExternalID(r, r.Identity.ID())
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}
	return transformers.UserToResponse(user), nil
}).
	WithSummary("Get my profile").
	WithTags("users").
	WithAuthentication().
	WithErrors(ErrUserNotFound)
