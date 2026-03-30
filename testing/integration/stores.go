//go:build integration

package integration

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opensearch-project/opensearch-go/v4"
	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/cereal"
	grubminio "github.com/zoobz-io/grub/minio"
	grubopensearch "github.com/zoobz-io/grub/opensearch"
	osrenderer "github.com/zoobz-io/lucene/opensearch"
	"github.com/zoobz-io/sum"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	admincontracts "github.com/zoobz-io/argus/admin/contracts"
	apicontracts "github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/wire"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest"
	"github.com/zoobz-io/argus/internal/vocabulary"
	"github.com/zoobz-io/argus/models"
	argusProto "github.com/zoobz-io/argus/proto"
	"github.com/zoobz-io/argus/stores"
)

// testStores is the shared store instance, initialized once in TestMain.
var testStores *stores.Stores

// InitStores creates the sum registry and stores in the correct order.
// Must be called once from TestMain before any tests run.
//
// Order matters: stores.New (which calls sum.NewDatabase) must run BEFORE
// sum.NewBoundary for the same types. NewBoundary mutates the sentinel type
// cache in a way that breaks subsequent NewDatabase calls (known sum/cereal bug).
func InitStores(ctx context.Context, infra *Infra) error {
	// Set up encryption for Provider credential store.encrypt:"aes" tag.
	encKey, _ := hex.DecodeString("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	enc, err := cereal.AES(encKey)
	if err != nil {
		return fmt.Errorf("creating AES encryptor: %w", err)
	}

	svc := sum.New()
	svc.WithEncryptor(cereal.EncryptAES, enc)
	k := sum.Start()

	db, err := sqlx.Connect("postgres", infra.PostgresDSN)
	if err != nil {
		return err
	}

	renderer := astqlpg.New()

	minioClient, err := minio.New(infra.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(infra.MinIOAccessKey, infra.MinIOSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}
	bucketProvider := grubminio.New(minioClient, "argus-test")

	osClient, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{infra.OpenSearchAddr},
	})
	if err != nil {
		return err
	}
	searchProvider := grubopensearch.New(osClient, grubopensearch.Config{
		Version: osrenderer.V2,
	})

	// Stores FIRST — creates sum.Database instances via grub.
	testStores = stores.New(db, renderer, bucketProvider, searchProvider)

	// Register contracts (so handlers can resolve them via sum.MustUse).
	sum.Register[apicontracts.Tenants](k, testStores.Tenants)
	sum.Register[apicontracts.Providers](k, testStores.Providers)
	sum.Register[apicontracts.WatchedPaths](k, testStores.WatchedPaths)
	sum.Register[apicontracts.Documents](k, testStores.Documents)
	sum.Register[apicontracts.DocumentVersions](k, testStores.DocumentVersions)
	sum.Register[apicontracts.DocumentVersionSearch](k, testStores.DocumentVersionSearch)
	sum.Register[apicontracts.Topics](k, testStores.Topics)
	sum.Register[apicontracts.Tags](k, testStores.Tags)
	sum.Register[apicontracts.Users](k, testStores.Users)
	sum.Register[apicontracts.Notifications](k, testStores.Notifications)
	sum.Register[apicontracts.Subscriptions](k, testStores.Subscriptions)
	sum.Register[apicontracts.JobReader](k, testStores.Jobs)
	sum.Register[apicontracts.Hooks](k, testStores.Hooks)
	sum.Register[apicontracts.Deliveries](k, testStores.Deliveries)
	sum.Register[apicontracts.IngestEnqueuer](k, ingest.NewEnqueuer())
	sum.Register[apicontracts.QueryEmbedder](k, &stubQueryEmbedder{dimensions: 1536})

	sum.Register[admincontracts.Tenants](k, testStores.Tenants)
	sum.Register[admincontracts.Providers](k, testStores.Providers)
	sum.Register[admincontracts.WatchedPaths](k, testStores.WatchedPaths)
	sum.Register[admincontracts.Documents](k, testStores.Documents)
	sum.Register[admincontracts.DocumentVersions](k, testStores.DocumentVersions)
	sum.Register[admincontracts.Topics](k, testStores.Topics)
	sum.Register[admincontracts.Tags](k, testStores.Tags)
	sum.Register[admincontracts.Users](k, testStores.Users)
	sum.Register[admincontracts.Subscriptions](k, testStores.AdminSubscriptions)
	sum.Register[admincontracts.Hooks](k, testStores.AdminHooks)
	sum.Register[admincontracts.AuditLog](k, testStores.Audit)

	// Internal contracts (pipeline).
	sum.Register[intcontracts.IngestVersions](k, testStores.DocumentVersions)
	sum.Register[intcontracts.IngestDocuments](k, testStores.Documents)
	sum.Register[intcontracts.IngestSearch](k, testStores.DocumentVersionSearch)
	sum.Register[intcontracts.IngestJobs](k, testStores.Jobs)
	sum.Register[intcontracts.IngestTopics](k, testStores.Topics)
	sum.Register[intcontracts.IngestTags](k, testStores.Tags)
	// OCR — real Tesseract sidecar via gRPC.
	ocrConn, err := grpc.NewClient(infra.OCRAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connecting to ocr: %w", err)
	}
	ocrClient := argusProto.NewOCRServiceClient(ocrConn)
	sum.Register[intcontracts.OCR](k, ocrClient)

	// Convert — real LibreOffice sidecar via gRPC.
	convertConn, err := grpc.NewClient(infra.ConvertAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("connecting to convert: %w", err)
	}
	convertClient := argusProto.NewConvertServiceClient(convertConn)
	sum.Register[intcontracts.Converter](k, convertClient)
	sum.Register[intcontracts.Classifier](k, &stubClassifier{})
	sum.Register[intcontracts.Analyzer](k, &stubAnalyzer{})
	sum.Register[intcontracts.Embedder](k, &stubEmbedder{dimensions: 1536})

	// Pipelines.
	vocabPipeline := vocabulary.New()
	sum.Register[apicontracts.Vocabulary](k, vocabPipeline)
	sum.Register[admincontracts.Vocabulary](k, vocabPipeline)

	// Boundaries AFTER stores.
	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.Provider](k)
	sum.NewBoundary[models.WatchedPath](k)
	sum.NewBoundary[models.Document](k)
	sum.NewBoundary[models.DocumentVersion](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Subscription](k)
	sum.NewBoundary[models.Job](k)
	sum.NewBoundary[models.Hook](k)
	sum.NewBoundary[models.Delivery](k)

	// Wire boundaries (needed for handler OnSend).
	wire.RegisterBoundaries(k)

	sum.Freeze(k)
	return nil
}

// Stores returns the shared store instance for tests.
func Stores(t *testing.T) *stores.Stores {
	t.Helper()
	if testStores == nil {
		t.Fatal("stores not initialized — call InitStores from TestMain")
	}
	return testStores
}
