package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listUsers = rocco.GET[rocco.NoBody, wire.AdminUserListResponse]("/users", func(r *rocco.Request[rocco.NoBody]) (wire.AdminUserListResponse, error) {
	store := sum.MustUse[contracts.Users](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListUsers(r, page)
	if err != nil {
		return wire.AdminUserListResponse{}, err
	}
	return transformers.UsersToAdminList(result), nil
}).
	WithSummary("List users").
	WithTags("users").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getUser = rocco.GET[rocco.NoBody, wire.AdminUserResponse]("/users/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminUserResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Users](r)
	user, err := store.GetUser(r, id)
	if err != nil {
		return wire.AdminUserResponse{}, ErrUserNotFound
	}
	return transformers.UserToAdminResponse(user), nil
}).
	WithSummary("Get user").
	WithTags("users").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

var createUser = rocco.POST[wire.AdminUserCreateRequest, wire.AdminUserResponse]("/users", func(r *rocco.Request[wire.AdminUserCreateRequest]) (wire.AdminUserResponse, error) {
	store := sum.MustUse[contracts.Users](r)
	user, err := store.CreateUser(r, r.Body.TenantID, r.Body.ExternalID, r.Body.Email, r.Body.DisplayName, r.Body.Role)
	if err != nil {
		return wire.AdminUserResponse{}, err
	}
	return transformers.UserToAdminResponse(user), nil
}).
	WithSummary("Create user").
	WithTags("users").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateUser = rocco.PUT[wire.AdminUserUpdateRequest, wire.AdminUserResponse]("/users/{id}", func(r *rocco.Request[wire.AdminUserUpdateRequest]) (wire.AdminUserResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Users](r)
	existing, err := store.GetUser(r, id)
	if err != nil {
		return wire.AdminUserResponse{}, ErrUserNotFound
	}
	email := existing.Email
	if r.Body.Email != nil {
		email = *r.Body.Email
	}
	displayName := existing.DisplayName
	if r.Body.DisplayName != nil {
		displayName = *r.Body.DisplayName
	}
	role := existing.Role
	if r.Body.Role != nil {
		role = *r.Body.Role
	}
	status := existing.Status
	if r.Body.Status != nil {
		status = *r.Body.Status
	}
	user, err := store.UpdateUser(r, id, email, displayName, role, status)
	if err != nil {
		return wire.AdminUserResponse{}, ErrUserNotFound
	}
	return transformers.UserToAdminResponse(user), nil
}).
	WithSummary("Update user").
	WithTags("users").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrUserNotFound, rocco.ErrValidationFailed)

var deleteUser = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/users/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Users](r)
	if err := store.DeleteUser(r, id); err != nil {
		return rocco.NoBody{}, ErrUserNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete user").
	WithTags("users").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrUserNotFound)
