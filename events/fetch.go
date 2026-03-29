package events

import "github.com/zoobz-io/capitan"

// Fetch signals for connector document download.
var (
	FetchSignal = capitan.NewSignal("argus.fetch", "Document fetch requested")
	FetchKey    = capitan.NewKey[FetchMessage]("fetch", "events.FetchMessage")
)

// FetchMessage is the payload emitted when a new document version needs fetching.
type FetchMessage struct {
	VersionID  string `json:"version_id"`
	DocumentID string `json:"document_id"`
	ProviderID string `json:"provider_id"`
	TenantID   string `json:"tenant_id"`
	Ref        string `json:"ref"`
	ObjectKey  string `json:"object_key"`
}

// Clone returns a deep copy.
func (m FetchMessage) Clone() FetchMessage {
	return m
}
