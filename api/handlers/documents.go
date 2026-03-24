package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listDocuments = rocco.GET[rocco.NoBody, wire.DocumentListResponse]("/documents", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentListResponse, error) {
	tid, err := tenantID(r.Identity)
	if err != nil {
		return wire.DocumentListResponse{}, rocco.ErrBadRequest.WithMessage("invalid tenant")
	}
	store := sum.MustUse[contracts.Documents](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListDocumentsByTenant(r, tid, page)
	if err != nil {
		return wire.DocumentListResponse{}, err
	}
	return transformers.DocumentsToListResponse(result, page.PageSize()), nil
}).
	WithSummary("List documents").
	WithTags("documents").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getDocument = rocco.GET[rocco.NoBody, wire.DocumentResponse]("/documents/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.DocumentResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Documents](r)
	doc, err := store.GetDocument(r, id)
	if err != nil {
		return wire.DocumentResponse{}, ErrDocumentNotFound
	}
	return transformers.DocumentToResponse(doc), nil
}).
	WithSummary("Get document").
	WithTags("documents").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrDocumentNotFound)
