package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	apiwire "github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/audit"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAllProviders = rocco.GET[rocco.NoBody, wire.AdminProviderListResponse]("/providers", func(r *rocco.Request[rocco.NoBody]) (wire.AdminProviderListResponse, error) {
	store := sum.MustUse[contracts.Providers](r)
	page := offsetPageFromQuery(r.Params)
	result, err := store.ListProviders(r, page)
	if err != nil {
		return wire.AdminProviderListResponse{}, err
	}
	return transformers.ProvidersToAdminList(result), nil
}).
	WithSummary("List all providers").
	WithTags("providers").
	WithQueryParams("offset", "limit").
	WithAuthentication()

var getAdminProvider = rocco.GET[rocco.NoBody, wire.AdminProviderResponse]("/providers/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminProviderResponse, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Providers](r)
	provider, err := store.GetProvider(r, id)
	if err != nil {
		return wire.AdminProviderResponse{}, ErrProviderNotFound
	}
	return transformers.ProviderToAdminResponse(provider), nil
}).
	WithSummary("Get provider").
	WithTags("providers").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrProviderNotFound)

var createAdminProvider = rocco.POST[apiwire.ProviderCreateRequest, wire.AdminProviderResponse]("/tenants/{tenant_id}/providers", func(r *rocco.Request[apiwire.ProviderCreateRequest]) (wire.AdminProviderResponse, error) {
	tid := pathID(r.Params, "tenant_id")
	store := sum.MustUse[contracts.Providers](r)
	provider, err := store.CreateProvider(r, tid, r.Body.Type, r.Body.Name, r.Body.Credentials)
	if err != nil {
		return wire.AdminProviderResponse{}, err
	}
	return transformers.ProviderToAdminResponse(provider), nil
}).
	WithSummary("Create provider for tenant").
	WithTags("providers").
	WithPathParams("tenant_id").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var deleteAdminProvider = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/providers/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id := pathID(r.Params, "id")
	store := sum.MustUse[contracts.Providers](r)
	if err := store.DeleteProvider(r, id); err != nil {
		return rocco.NoBody{}, ErrProviderNotFound
	}
	audit.Emit(r, "provider.deleted", "provider", id, "", r.Identity.ID(), nil)
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete provider").
	WithTags("providers").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrProviderNotFound)
