package boot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/vex"
	vexopenai "github.com/zoobz-io/vex/openai"
	"github.com/zoobz-io/zyn"
	zynopenai "github.com/zoobz-io/zyn/openai"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/internal/ingest"
	"github.com/zoobz-io/argus/models"
)

// LLM creates a document analyzer synapse from config.
func LLM(ctx context.Context) (*ingest.SynapseAnalyzer, error) {
	cfg := sum.MustUse[config.LLM](ctx)
	provider := zynopenai.New(zynopenai.Config{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
	})
	synapse, err := zyn.Extract[models.DocumentAnalysis](
		"Analyze the document content and extract: a concise summary paragraph, the ISO 639-1 language code, and any matching topics and tags from the provided vocabulary lists. Only select topics and tags that clearly apply to the content.",
		provider,
		zyn.WithRetry(3),
		zyn.WithTimeout(60*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("creating analyzer synapse: %w", err)
	}
	analyzer := ingest.NewSynapseAnalyzer(synapse)
	log.Println("llm provider initialized")
	return analyzer, nil
}

// Embedding creates a vector embedding service from config.
func Embedding(ctx context.Context) (*vex.Service, error) {
	cfg := sum.MustUse[config.Embedding](ctx)
	provider := vexopenai.New(vexopenai.Config{
		APIKey:     cfg.APIKey,
		Model:      cfg.Model,
		BaseURL:    cfg.BaseURL,
		Dimensions: cfg.Dimensions,
	})
	service := vex.NewService(provider,
		vex.WithRetry(3),
		vex.WithTimeout(30*time.Second),
	)
	log.Println("embedding provider initialized")
	return service, nil
}
