package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listTags = rocco.GET[rocco.NoBody, wire.TagListResponse]("/tags", func(r *rocco.Request[rocco.NoBody]) (wire.TagListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Tags](r)
	tags, err := store.ListTagsByTenant(r, tid)
	if err != nil {
		return wire.TagListResponse{}, err
	}
	return transformers.TagsToListResponse(tags), nil
}).
	WithSummary("List tags").
	WithTags("tags").
	WithAuthentication()

var getTag = rocco.GET[rocco.NoBody, wire.TagResponse]("/tags/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.TagResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Tags](r)
	tag, err := store.GetTag(r, id)
	if err != nil {
		return wire.TagResponse{}, ErrTagNotFound
	}
	return transformers.TagToResponse(tag), nil
}).
	WithSummary("Get tag").
	WithTags("tags").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTagNotFound)

var createTag = rocco.POST[wire.TagCreateRequest, wire.TagResponse]("/tags", func(r *rocco.Request[wire.TagCreateRequest]) (wire.TagResponse, error) {
	tid := tenantID(r.Identity)
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.Process(r, tid, r.Body.Name, r.Body.Description); err != nil {
		return wire.TagResponse{}, err
	}
	store := sum.MustUse[contracts.Tags](r)
	tag, err := store.CreateTag(r, tid, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.TagResponse{}, err
	}
	return transformers.TagToResponse(tag), nil
}).
	WithSummary("Create tag").
	WithTags("tags").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateTag = rocco.PUT[wire.TagCreateRequest, wire.TagResponse]("/tags/{id}", func(r *rocco.Request[wire.TagCreateRequest]) (wire.TagResponse, error) {
	id := pathID(r.Params, "id")
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.ProcessUpdate(r, id, r.Body.Name, r.Body.Description); err != nil {
		return wire.TagResponse{}, err
	}
	store := sum.MustUse[contracts.Tags](r)
	tag, err := store.UpdateTag(r, id, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.TagResponse{}, ErrTagNotFound
	}
	return transformers.TagToResponse(tag), nil
}).
	WithSummary("Update tag").
	WithTags("tags").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTagNotFound, rocco.ErrValidationFailed)
