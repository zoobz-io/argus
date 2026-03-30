package handlers

import (
	"strconv"
	"time"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var listAuditLog = rocco.GET[rocco.NoBody, wire.AuditListResponse]("/audit", func(r *rocco.Request[rocco.NoBody]) (wire.AuditListResponse, error) {
	tid := tenantID(r.Identity)
	store := sum.MustUse[contracts.AuditLog](r)

	params := auditSearchFromQuery(r.Params)
	params.TenantID = tid

	result, err := store.Search(r, params)
	if err != nil {
		return wire.AuditListResponse{}, err
	}
	return transformers.DomainEventsToListResponse(result), nil
}).
	WithSummary("List audit log").
	WithTags("audit").
	WithQueryParams("action", "resource_type", "actor_id", "from", "to", "offset", "limit").
	WithAuthentication()

func auditSearchFromQuery(params *rocco.Params) models.DomainEventSearchParams {
	p := models.DomainEventSearchParams{
		Limit: models.DefaultPageSize,
	}
	if v := params.Query["action"]; v != "" {
		p.Action = v
	}
	if v := params.Query["resource_type"]; v != "" {
		p.ResourceType = v
	}
	if v := params.Query["actor_id"]; v != "" {
		p.ActorID = v
	}
	if v := params.Query["from"]; v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			p.From = &t
		}
	}
	if v := params.Query["to"]; v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			p.To = &t
		}
	}
	if v := params.Query["offset"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p.Offset = n
		}
	}
	if v := params.Query["limit"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p.Limit = n
		}
	}
	return p
}
