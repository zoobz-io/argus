package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listDocumentVersions = rocco.GET[rocco.NoBody, wire.DocumentVersionListResponse]("/documents/{document_id}/versions", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentVersionListResponse, error) {
	docID, err := pathID(r.Params, "document_id")
	if err != nil {
		return wire.DocumentVersionListResponse{}, rocco.ErrBadRequest.WithMessage("invalid document_id")
	}
	store := sum.MustUse[contracts.DocumentVersions](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListVersionsByDocument(r, docID, page)
	if err != nil {
		return wire.DocumentVersionListResponse{}, err
	}
	return transformers.DocumentVersionsToListResponse(result, page.PageSize()), nil
}).
	WithSummary("List document versions").
	WithTags("document-versions").
	WithPathParams("document_id").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getDocumentVersion = rocco.GET[rocco.NoBody, wire.DocumentVersionResponse]("/documents/{document_id}/versions/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentVersionResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.DocumentVersionResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.DocumentVersions](r)
	version, err := store.GetDocumentVersion(r, id)
	if err != nil {
		return wire.DocumentVersionResponse{}, ErrVersionNotFound
	}
	return transformers.DocumentVersionToResponse(version), nil
}).
	WithSummary("Get document version").
	WithTags("document-versions").
	WithPathParams("document_id", "id").
	WithAuthentication().
	WithErrors(ErrVersionNotFound)
