package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listTopics = rocco.GET[rocco.NoBody, wire.TopicListResponse]("/topics", func(r *rocco.Request[rocco.NoBody]) (wire.TopicListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Topics](r)
	topics, err := store.ListTopicsByTenant(r, tid)
	if err != nil {
		return wire.TopicListResponse{}, err
	}
	return transformers.TopicsToListResponse(topics), nil
}).
	WithSummary("List topics").
	WithTags("topics").
	WithAuthentication()

var getTopic = rocco.GET[rocco.NoBody, wire.TopicResponse]("/topics/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.TopicResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Topics](r)
	topic, err := store.GetTopic(r, id)
	if err != nil {
		return wire.TopicResponse{}, ErrTopicNotFound
	}
	return transformers.TopicToResponse(topic), nil
}).
	WithSummary("Get topic").
	WithTags("topics").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTopicNotFound)

var createTopic = rocco.POST[wire.TopicCreateRequest, wire.TopicResponse]("/topics", func(r *rocco.Request[wire.TopicCreateRequest]) (wire.TopicResponse, error) {
	tid := tenantID(r.Identity)
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.Process(r, tid, r.Body.Name, r.Body.Description); err != nil {
		return wire.TopicResponse{}, err
	}
	store := sum.MustUse[contracts.Topics](r)
	topic, err := store.CreateTopic(r, tid, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.TopicResponse{}, err
	}
	return transformers.TopicToResponse(topic), nil
}).
	WithSummary("Create topic").
	WithTags("topics").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateTopic = rocco.PUT[wire.TopicCreateRequest, wire.TopicResponse]("/topics/{id}", func(r *rocco.Request[wire.TopicCreateRequest]) (wire.TopicResponse, error) {
	id := pathID(r.Params, "id")
	vocab := sum.MustUse[contracts.Vocabulary](r)
	if err := vocab.ProcessUpdate(r, id, r.Body.Name, r.Body.Description); err != nil {
		return wire.TopicResponse{}, err
	}
	store := sum.MustUse[contracts.Topics](r)
	topic, err := store.UpdateTopic(r, id, r.Body.Name, r.Body.Description)
	if err != nil {
		return wire.TopicResponse{}, ErrTopicNotFound
	}
	return transformers.TopicToResponse(topic), nil
}).
	WithSummary("Update topic").
	WithTags("topics").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTopicNotFound, rocco.ErrValidationFailed)
