package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAllDocuments = rocco.GET[rocco.NoBody, wire.AdminDocumentListResponse]("/documents", func(r *rocco.Request[rocco.NoBody]) (wire.AdminDocumentListResponse, error) {
	store := sum.MustUse[contracts.Documents](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListDocuments(r, page)
	if err != nil {
		return wire.AdminDocumentListResponse{}, err
	}
	return transformers.DocumentsToAdminList(result, page.PageSize()), nil
}).
	WithSummary("List all documents").
	WithTags("documents").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getAdminDocument = rocco.GET[rocco.NoBody, wire.AdminDocumentResponse]("/documents/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminDocumentResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.AdminDocumentResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Documents](r)
	doc, err := store.GetDocument(r, id)
	if err != nil {
		return wire.AdminDocumentResponse{}, ErrDocumentNotFound
	}
	return transformers.DocumentToAdminResponse(doc), nil
}).
	WithSummary("Get document").
	WithTags("documents").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrDocumentNotFound)

var deleteAdminDocument = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/documents/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return rocco.NoBody{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Documents](r)
	if err := store.DeleteDocument(r, id); err != nil {
		return rocco.NoBody{}, ErrDocumentNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete document").
	WithTags("documents").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrDocumentNotFound)
