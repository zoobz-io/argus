package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var searchDocuments = rocco.POST[wire.SearchRequest, wire.SearchResponse]("/search", func(r *rocco.Request[wire.SearchRequest]) (wire.SearchResponse, error) {
	if r.Body.Query == "" {
		return wire.SearchResponse{}, ErrMissingQuery
	}
	store := sum.MustUse[contracts.DocumentVersionSearch](r)
	result, err := store.Search(r, r.Body.Query, r.Body.Size, r.Body.From)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	return transformers.SearchResultToResponse(result), nil
}).
	WithSummary("Search documents").
	WithTags("search").
	WithAuthentication().
	WithErrors(ErrMissingQuery)
