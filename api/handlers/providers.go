package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/audit"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listProviders = rocco.GET[rocco.NoBody, wire.ProviderListResponse]("/providers", func(r *rocco.Request[rocco.NoBody]) (wire.ProviderListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Providers](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListProvidersByTenant(r, tid, page)
	if err != nil {
		return wire.ProviderListResponse{}, err
	}
	return transformers.ProvidersToListResponse(result), nil
}).
	WithSummary("List providers").
	WithTags("providers").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getProvider = rocco.GET[rocco.NoBody, wire.ProviderResponse]("/providers/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.ProviderResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Providers](r)
	provider, err := store.GetProvider(r, id)
	if err != nil {
		return wire.ProviderResponse{}, ErrProviderNotFound
	}
	return transformers.ProviderToResponse(provider), nil
}).
	WithSummary("Get provider").
	WithTags("providers").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrProviderNotFound)

var createProvider = rocco.POST[wire.ProviderCreateRequest, wire.ProviderResponse]("/providers", func(r *rocco.Request[wire.ProviderCreateRequest]) (wire.ProviderResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.Providers](r)
	provider, err := store.CreateProvider(r, tid, r.Body.Type, r.Body.Name, r.Body.Credentials)
	if err != nil {
		return wire.ProviderResponse{}, err
	}
	audit.Emit(r, "provider.created", "provider", provider.ID, tid, r.Identity.ID(), map[string]any{
		"provider_type": string(r.Body.Type),
		"provider_name": r.Body.Name,
	})
	return transformers.ProviderToResponse(provider), nil
}).
	WithSummary("Create provider").
	WithTags("providers").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateProvider = rocco.PUT[wire.ProviderCreateRequest, wire.ProviderResponse]("/providers/{id}", func(r *rocco.Request[wire.ProviderCreateRequest]) (wire.ProviderResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Providers](r)
	provider, err := store.UpdateProvider(r, id, r.Body.Type, r.Body.Name, r.Body.Credentials)
	if err != nil {
		return wire.ProviderResponse{}, ErrProviderNotFound
	}
	audit.Emit(r, "provider.updated", "provider", provider.ID, tenantID(r.Identity), r.Identity.ID(), map[string]any{
		"provider_type": string(r.Body.Type),
		"provider_name": r.Body.Name,
	})
	return transformers.ProviderToResponse(provider), nil
}).
	WithSummary("Update provider").
	WithTags("providers").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrProviderNotFound, rocco.ErrValidationFailed)
