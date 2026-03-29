package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/audit"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listWatchedPaths = rocco.GET[rocco.NoBody, wire.WatchedPathListResponse]("/watched-paths", func(r *rocco.Request[rocco.NoBody]) (wire.WatchedPathListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.WatchedPaths](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListWatchedPathsByTenant(r, tid, page)
	if err != nil {
		return wire.WatchedPathListResponse{}, err
	}
	return transformers.WatchedPathsToListResponse(result), nil
}).
	WithSummary("List watched paths").
	WithTags("watched-paths").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getWatchedPath = rocco.GET[rocco.NoBody, wire.WatchedPathResponse]("/watched-paths/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.WatchedPathResponse, error) {
	id := pathID(r.Params, "id")
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
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.CreateWatchedPath(r, tid, r.Body.ProviderID, r.Body.Path)
	if err != nil {
		return wire.WatchedPathResponse{}, err
	}
	audit.Emit(r, "watched_path.created", "watched_path", wp.ID, tid, r.Identity.ID(), map[string]any{
		"provider_id": r.Body.ProviderID,
		"path":        r.Body.Path,
	})
	return transformers.WatchedPathToResponse(wp), nil
}).
	WithSummary("Create watched path").
	WithTags("watched-paths").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateWatchedPath = rocco.PUT[wire.WatchedPathCreateRequest, wire.WatchedPathResponse]("/watched-paths/{id}", func(r *rocco.Request[wire.WatchedPathCreateRequest]) (wire.WatchedPathResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.UpdateWatchedPath(r, id, r.Body.Path)
	if err != nil {
		return wire.WatchedPathResponse{}, ErrWatchedPathNotFound
	}
	audit.Emit(r, "watched_path.updated", "watched_path", wp.ID, tenantID(r.Identity), r.Identity.ID(), map[string]any{
		"path": r.Body.Path,
	})
	return transformers.WatchedPathToResponse(wp), nil
}).
	WithSummary("Update watched path").
	WithTags("watched-paths").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrWatchedPathNotFound, rocco.ErrValidationFailed)
