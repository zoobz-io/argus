package events

import "github.com/zoobz-io/capitan"

// Ingestion queue signals for async job dispatch.
var (
	IngestQueueSignal = capitan.NewSignal("argus.ingest.queue", "Ingestion job queued for worker processing")
	IngestQueueKey    = capitan.NewKey[IngestMessage]("ingest_queue", "events.IngestMessage")
)

// IngestMessage is the payload published to the ingestion herald stream.
type IngestMessage struct {
	JobID     string `json:"job_id"`
	VersionID string `json:"version_id"`
}

// Clone returns a deep copy.
func (m IngestMessage) Clone() IngestMessage {
	return m
}
