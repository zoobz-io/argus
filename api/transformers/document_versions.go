// Package transformers provides functions to convert domain models into public API wire types.
package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionToResponse transforms a DocumentVersion model to an API response.
func DocumentVersionToResponse(v *models.DocumentVersion) wire.DocumentVersionResponse {
	return wire.DocumentVersionResponse{
		ID:               v.ID,
		DocumentID:       v.DocumentID,
		VersionNumber:    v.VersionNumber,
		ExtractionStatus: v.ExtractionStatus,
	}
}

// DocumentVersionsToResponse transforms a slice of DocumentVersion models to responses.
func DocumentVersionsToResponse(versions []*models.DocumentVersion) []wire.DocumentVersionResponse {
	result := make([]wire.DocumentVersionResponse, len(versions))
	for i, v := range versions {
		result[i] = DocumentVersionToResponse(v)
	}
	return result
}
