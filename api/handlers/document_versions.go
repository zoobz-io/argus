package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listDocumentVersions = rocco.GET[rocco.NoBody, wire.DocumentVersionListResponse]("/documents/{document_id}/versions", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentVersionListResponse, error) {
	docID := pathID(r.Params, "document_id")
	store := sum.MustUse[contracts.DocumentVersions](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListVersionsByDocument(r, docID, page)
	if err != nil {
		return wire.DocumentVersionListResponse{}, err
	}
	return transformers.DocumentVersionsToListResponse(result), nil
}).
	WithSummary("List document versions").
	WithTags("document-versions").
	WithPathParams("document_id").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getDocumentVersion = rocco.GET[rocco.NoBody, wire.DocumentVersionResponse]("/documents/{document_id}/versions/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.DocumentVersionResponse, error) {
	id := pathID(r.Params, "id")
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
