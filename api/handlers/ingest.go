package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/event"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

var triggerIngest = rocco.POST[wire.IngestRequest, wire.IngestResponse]("/ingest", func(r *rocco.Request[wire.IngestRequest]) (wire.IngestResponse, error) {
	tid := tenantID(r.Identity)
	enqueuer := sum.MustUse[contracts.IngestEnqueuer](r)
	job, err := enqueuer.Enqueue(r, r.Body.VersionID, tid)
	if err != nil {
		return wire.IngestResponse{}, err
	}
	event.Emit(r, "document.ingested", "Document queued for ingestion", "document", job.ID, tid, r.Identity.ID(), map[string]any{
		"version_id": r.Body.VersionID,
	})
	return wire.IngestResponse{
		JobID:  job.ID,
		Status: string(job.Status),
	}, nil
}).
	WithSummary("Trigger async ingestion").
	WithTags("ingestion").
	WithAuthentication()
