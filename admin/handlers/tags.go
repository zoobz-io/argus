package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	apiwire "github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAdminTags = rocco.GET[rocco.NoBody, wire.AdminTagListResponse]("/tenants/{tenant_id}/tags", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTagListResponse, error) {
	tid := pathID(r.Params, "tenant_id")
	store := sum.MustUse[contracts.Tags](r)
	tags, err := store.ListTagsByTenant(r, tid)
	if err != nil {
		return wire.AdminTagListResponse{}, err
	}
	return transformers.TagsToAdminList(tags), nil
}).
	WithSummary("List tags for tenant").
	WithTags("tags").
	WithPathParams("tenant_id").
	WithAuthentication()

var getAdminTag = rocco.GET[rocco.NoBody, wire.AdminTagResponse]("/tags/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTagResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Tags](r)
	tag, err := store.GetTag(r, id)
	if err != nil {
		return wire.AdminTagResponse{}, ErrTagNotFound
	}
	return transformers.TagToAdminResponse(tag), nil
}).
	WithSummary("Get tag").
	WithTags("tags").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTagNotFound)

var createAdminTag = rocco.POST[apiwire.TagCreateRequest, wire.AdminTagResponse]("/tenants/{tenant_id}/tags", func(r *rocco.Request[apiwire.TagCreateRequest]) (wire.AdminTagResponse, error) {
	tid := pathID(r.Params, "tenant_id")
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.Process(r, tid, r.Body.Name, r.Body.Description); err != nil {
		return wire.AdminTagResponse{}, err
	}
	store := sum.MustUse[contracts.Tags](r)
	tag, err := store.CreateTag(r, tid, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.AdminTagResponse{}, err
	}
	return transformers.TagToAdminResponse(tag), nil
}).
	WithSummary("Create tag for tenant").
	WithTags("tags").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var deleteAdminTag = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tags/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Tags](r)
	if err := store.DeleteTag(r, id); err != nil {
		return rocco.NoBody{}, ErrTagNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete tag").
	WithTags("tags").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrTagNotFound)
