package ingest

import (
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/vex"
)

// DocumentContext carries data through the ingestion pipeline stages.
type DocumentContext struct {
	Version  *models.DocumentVersion
	Document *models.Document
	RawBytes []byte
	Content  string
	Summary  string
	Embedding vex.Vector
}

// Clone returns a deep copy of the document context for pipz compatibility.
func (dc *DocumentContext) Clone() *DocumentContext {
	c := &DocumentContext{
		Content: dc.Content,
		Summary: dc.Summary,
	}
	if dc.Version != nil {
		v := dc.Version.Clone()
		c.Version = &v
	}
	if dc.Document != nil {
		d := dc.Document.Clone()
		c.Document = &d
	}
	if dc.RawBytes != nil {
		c.RawBytes = make([]byte, len(dc.RawBytes))
		copy(c.RawBytes, dc.RawBytes)
	}
	if dc.Embedding != nil {
		c.Embedding = make(vex.Vector, len(dc.Embedding))
		copy(c.Embedding, dc.Embedding)
	}
	return c
}
