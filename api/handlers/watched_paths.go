package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listWatchedPaths = rocco.GET[rocco.NoBody, wire.WatchedPathListResponse]("/watched-paths", func(r *rocco.Request[rocco.NoBody]) (wire.WatchedPathListResponse, error) {
	tid, err := tenantID(r.Identity)
	if err != nil {
		return wire.WatchedPathListResponse{}, rocco.ErrBadRequest.WithMessage("invalid tenant")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListWatchedPathsByTenant(r, tid, page)
	if err != nil {
		return wire.WatchedPathListResponse{}, err
	}
	return transformers.WatchedPathsToListResponse(result, page.PageSize()), nil
}).
	WithSummary("List watched paths").
	WithTags("watched-paths").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getWatchedPath = rocco.GET[rocco.NoBody, wire.WatchedPathResponse]("/watched-paths/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.WatchedPathResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.WatchedPathResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.GetWatchedPath(r, id)
	if err != nil {
		return wire.WatchedPathResponse{}, ErrWatchedPathNotFound
	}
	return transformers.WatchedPathToResponse(wp), nil
}).
	WithSummary("Get watched path").
	WithTags("watched-paths").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrWatchedPathNotFound)

var createWatchedPath = rocco.POST[wire.WatchedPathCreateRequest, wire.WatchedPathResponse]("/watched-paths", func(r *rocco.Request[wire.WatchedPathCreateRequest]) (wire.WatchedPathResponse, error) {
	tid, err := tenantID(r.Identity)
	if err != nil {
		return wire.WatchedPathResponse{}, rocco.ErrBadRequest.WithMessage("invalid tenant")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.CreateWatchedPath(r, tid, r.Body.ProviderID, r.Body.Path)
	if err != nil {
		return wire.WatchedPathResponse{}, err
	}
	return transformers.WatchedPathToResponse(wp), nil
}).
	WithSummary("Create watched path").
	WithTags("watched-paths").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateWatchedPath = rocco.PUT[wire.WatchedPathCreateRequest, wire.WatchedPathResponse]("/watched-paths/{id}", func(r *rocco.Request[wire.WatchedPathCreateRequest]) (wire.WatchedPathResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.WatchedPathResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.UpdateWatchedPath(r, id, r.Body.Path)
	if err != nil {
		return wire.WatchedPathResponse{}, ErrWatchedPathNotFound
	}
	return transformers.WatchedPathToResponse(wp), nil
}).
	WithSummary("Update watched path").
	WithTags("watched-paths").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrWatchedPathNotFound, rocco.ErrValidationFailed)
