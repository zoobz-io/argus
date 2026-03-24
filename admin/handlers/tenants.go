package handlers

import (
	"github.com/zoobz-io/argus/admin/contracts"
	"github.com/zoobz-io/argus/admin/transformers"
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listTenants = rocco.GET[rocco.NoBody, wire.AdminTenantListResponse]("/tenants", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTenantListResponse, error) {
	store := sum.MustUse[contracts.Tenants](r)
	page := cursorPageFromQuery(r.Params)
	result, err := store.ListTenants(r, page)
	if err != nil {
		return wire.AdminTenantListResponse{}, err
	}
	return transformers.TenantsToAdminList(result, page.PageSize()), nil
}).
	WithSummary("List tenants").
	WithTags("tenants").
	WithQueryParams("cursor", "limit").
	WithAuthentication()

var getTenant = rocco.GET[rocco.NoBody, wire.AdminTenantResponse]("/tenants/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.AdminTenantResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.AdminTenantResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Tenants](r)
	tenant, err := store.GetTenant(r, id)
	if err != nil {
		return wire.AdminTenantResponse{}, ErrTenantNotFound
	}
	return transformers.TenantToAdminResponse(tenant), nil
}).
	WithSummary("Get tenant").
	WithTags("tenants").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound)

var createTenant = rocco.POST[wire.AdminTenantCreateRequest, wire.AdminTenantResponse]("/tenants", func(r *rocco.Request[wire.AdminTenantCreateRequest]) (wire.AdminTenantResponse, error) {
	store := sum.MustUse[contracts.Tenants](r)
	tenant, err := store.CreateTenant(r, r.Body.Name, r.Body.Slug)
	if err != nil {
		return wire.AdminTenantResponse{}, err
	}
	return transformers.TenantToAdminResponse(tenant), nil
}).
	WithSummary("Create tenant").
	WithTags("tenants").
	WithSuccessStatus(201).
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var updateTenant = rocco.PUT[wire.AdminTenantCreateRequest, wire.AdminTenantResponse]("/tenants/{id}", func(r *rocco.Request[wire.AdminTenantCreateRequest]) (wire.AdminTenantResponse, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return wire.AdminTenantResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Tenants](r)
	tenant, err := store.UpdateTenant(r, id, r.Body.Name, r.Body.Slug)
	if err != nil {
		return wire.AdminTenantResponse{}, ErrTenantNotFound
	}
	return transformers.TenantToAdminResponse(tenant), nil
}).
	WithSummary("Update tenant").
	WithTags("tenants").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrTenantNotFound, rocco.ErrValidationFailed)

var deleteTenant = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/tenants/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	id, err := pathID(r.Params, "id")
	if err != nil {
		return rocco.NoBody{}, rocco.ErrBadRequest.WithMessage("invalid id")
	}
	store := sum.MustUse[contracts.Tenants](r)
	if err := store.DeleteTenant(r, id); err != nil {
		return rocco.NoBody{}, ErrTenantNotFound
	}
	return rocco.NoBody{}, nil
}).
	WithSummary("Delete tenant").
	WithTags("tenants").
	WithPathParams("id").
	WithSuccessStatus(204).
	WithAuthentication().
	WithErrors(ErrTenantNotFound)
