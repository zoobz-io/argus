package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAllWatchedPaths = rocco.GET[rocco.NoBody, wire.AdminWatchedPathListResponse]("/watched-paths", func(r *rocco.Request[rocco.NoBody]) (wire.AdminWatchedPathListResponse, error) {
	store := sum.MustUse[contracts.WatchedPaths](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListWatchedPaths(r, page)
	if err != nil {
		return wire.AdminWatchedPathListResponse{}, err
	}
	return transformers.WatchedPathsToAdminList(result, page.PageSize()), nil
}).
	WithSummary("List all watched paths").
	WithTags("watched-paths").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getAdminWatchedPath = rocco.GET[rocco.NoBody, wire.AdminWatchedPathResponse]("/watched-paths/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminWatchedPathResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.AdminWatchedPathResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	wp, err := store.GetWatchedPath(r, id)
	if err != nil {
		return wire.AdminWatchedPathResponse{}, ErrWatchedPathNotFound
	}
	return transformers.WatchedPathToAdminResponse(wp), nil
}).
	WithSummary("Get watched path").
	WithTags("watched-paths").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrWatchedPathNotFound)

var deleteAdminWatchedPath = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/watched-paths/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return rocco.NoBody{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.WatchedPaths](r)
	if err := store.DeleteWatchedPath(r, id); err != nil {
		return rocco.NoBody{}, ErrWatchedPathNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete watched path").
	WithTags("watched-paths").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrWatchedPathNotFound)
