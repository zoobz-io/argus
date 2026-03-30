// Package opensearch embeds the OpenSearch index mapping files.
package opensearch

import "embed"

// Mappings contains all JSON mapping files for OpenSearch indices.
//
//go:embed *.json
var Mappings embed.FS
