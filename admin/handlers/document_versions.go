package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAllDocumentVersions = rocco.GET[rocco.NoBody, wire.AdminDocumentVersionListResponse]("/document-versions", func(r *rocco.Request[rocco.NoBody]) (wire.AdminDocumentVersionListResponse, error) {
	store := sum.MustUse[contracts.DocumentVersions](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListDocumentVersions(r, page)
	if err != nil {
		return wire.AdminDocumentVersionListResponse{}, err
	}
	return transformers.DocumentVersionsToAdminList(result, page.PageSize()), nil
}).
	WithSummary("List all document versions").
	WithTags("document-versions").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getAdminDocumentVersion = rocco.GET[rocco.NoBody, wire.AdminDocumentVersionResponse]("/document-versions/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminDocumentVersionResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.AdminDocumentVersionResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.DocumentVersions](r)
	version, err := store.GetDocumentVersion(r, id)
	if err != nil {
		return wire.AdminDocumentVersionResponse{}, ErrVersionNotFound
	}
	return transformers.DocumentVersionToAdminResponse(version), nil
}).
	WithSummary("Get document version").
	WithTags("document-versions").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrVersionNotFound)

var deleteAdminDocumentVersion = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/document-versions/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return rocco.NoBody{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.DocumentVersions](r)
	if err := store.DeleteDocumentVersion(r, id); err != nil {
		return rocco.NoBody{}, ErrVersionNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete document version").
	WithTags("document-versions").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrVersionNotFound)
