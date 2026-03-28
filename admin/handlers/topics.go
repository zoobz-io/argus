package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	apiwire "github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAdminTopics = rocco.GET[rocco.NoBody, wire.AdminTopicListResponse]("/tenants/{tenant_id}/topics", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTopicListResponse, error) {
	tid := pathID(r.Params, "tenant_id")
	store := sum.MustUse[contracts.Topics](r)
	topics, err := store.ListTopicsByTenant(r, tid)
	if err != nil {
		return wire.AdminTopicListResponse{}, err
	}
	return transformers.TopicsToAdminList(topics), nil
}).
	WithSummary("List topics for tenant").
	WithTags("topics").
	WithPathParams("tenant_id").
	WithAuthentication()

var getAdminTopic = rocco.GET[rocco.NoBody, wire.AdminTopicResponse]("/topics/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTopicResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Topics](r)
	topic, err := store.GetTopic(r, id)
	if err != nil {
		return wire.AdminTopicResponse{}, ErrTopicNotFound
	}
	return transformers.TopicToAdminResponse(topic), nil
}).
	WithSummary("Get topic").
	WithTags("topics").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTopicNotFound)

var createAdminTopic = rocco.POST[apiwire.TopicCreateRequest, wire.AdminTopicResponse]("/tenants/{tenant_id}/topics", func(r *rocco.Request[apiwire.TopicCreateRequest]) (wire.AdminTopicResponse, error) {
	tid := pathID(r.Params, "tenant_id")
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.Process(r, tid, r.Body.Name, r.Body.Description); err != nil {
		return wire.AdminTopicResponse{}, err
	}
	store := sum.MustUse[contracts.Topics](r)
	topic, err := store.CreateTopic(r, tid, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.AdminTopicResponse{}, err
	}
	return transformers.TopicToAdminResponse(topic), nil
}).
	WithSummary("Create topic for tenant").
	WithTags("topics").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var deleteAdminTopic = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/topics/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Topics](r)
	if err := store.DeleteTopic(r, id); err != nil {
		return rocco.NoBody{}, ErrTopicNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete topic").
	WithTags("topics").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrTopicNotFound)
