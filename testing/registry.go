//go:build testing

package argustest

import (
	"context"
	"net/http"
	"testing"

	admincontracts "github.com/zoobz-io/argus/admin/contracts"
	apicontracts "github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/config"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/oauth"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
	"github.com/zoobz-io/rocco"
	rtesting "github.com/zoobz-io/rocco/testing"
	"github.com/zoobz-io/sum"
)

// RegistryOption configures a mock contract in the sum registry.
type RegistryOption func(k sum.Key)

// API contract options.

func WithAPITenants(m *MockTenants) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Tenants](k, m) }
}
func WithAPIProviders(m *MockProviders) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Providers](k, m) }
}
func WithAPIWatchedPaths(m *MockWatchedPaths) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.WatchedPaths](k, m) }
}
func WithAPIDocuments(m *MockDocuments) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Documents](k, m) }
}
func WithAPIDocumentVersions(m *MockDocumentVersions) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.DocumentVersions](k, m) }
}
func WithAPIDocumentVersionSearch(m *MockDocumentVersionSearch) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.DocumentVersionSearch](k, m) }
}
func WithAPITopics(m *MockTopics) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Topics](k, m) }
}
func WithAPITags(m *MockTags) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Tags](k, m) }
}
func WithAPIIngest(m *MockIngest) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Ingest](k, m) }
}
func WithAPIIngestEnqueuer(m *MockIngestEnqueuer) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.IngestEnqueuer](k, m) }
}
func WithAPIJobReader(m *MockJobReader) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.JobReader](k, m) }
}
func WithAPIQueryEmbedder(m *MockQueryEmbedder) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.QueryEmbedder](k, m) }
}
func WithAPIVocabulary(m *MockVocabulary) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Vocabulary](k, m) }
}
// WithRegistration registers an arbitrary value by calling the provided function.
// Used for types that can't be imported without creating cycles.
func WithRegistration(fn func(sum.Key)) RegistryOption {
	return func(k sum.Key) { fn(k) }
}
func WithProviderRegistry(reg *provider.Registry) RegistryOption {
	return func(k sum.Key) { sum.Register[*provider.Registry](k, reg) }
}
func WithStateSigner(signer *oauth.StateSigner) RegistryOption {
	return func(k sum.Key) { sum.Register[*oauth.StateSigner](k, signer) }
}
func WithAPIUsers(m *MockUsers) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Users](k, m) }
}
func WithAPISubscriptions(m *MockSubscriptions) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Subscriptions](k, m) }
}
func WithAPINotifications(m *MockNotifications) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.Notifications](k, m) }
}
func WithAPIAuditLog(m *MockAuditLog) RegistryOption {
	return func(k sum.Key) { sum.Register[apicontracts.AuditLog](k, m) }
}

// Admin contract options.

func WithAdminTenants(m *MockTenants) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Tenants](k, m) }
}
func WithAdminProviders(m *MockProviders) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Providers](k, m) }
}
func WithAdminWatchedPaths(m *MockWatchedPaths) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.WatchedPaths](k, m) }
}
func WithAdminDocuments(m *MockDocuments) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Documents](k, m) }
}
func WithAdminDocumentVersions(m *MockDocumentVersions) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.DocumentVersions](k, m) }
}
func WithAdminTopics(m *MockTopics) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Topics](k, m) }
}
func WithAdminTags(m *MockTags) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Tags](k, m) }
}
func WithAdminVocabulary(m *MockVocabulary) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Vocabulary](k, m) }
}
func WithAdminUsers(m *MockUsers) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Users](k, m) }
}
func WithAdminSubscriptions(m *MockAdminSubscriptions) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.Subscriptions](k, m) }
}
func WithAdminAuditLog(m *MockAdminAuditLog) RegistryOption {
	return func(k sum.Key) { sum.Register[admincontracts.AuditLog](k, m) }
}

// Internal contract options.

func WithOCR(m *MockOCR) RegistryOption {
	return func(k sum.Key) { sum.Register[intcontracts.OCR](k, m) }
}
func WithConverter(m *MockConverter) RegistryOption {
	return func(k sum.Key) { sum.Register[intcontracts.Converter](k, m) }
}
func WithClassifier(m *MockClassifier) RegistryOption {
	return func(k sum.Key) { sum.Register[intcontracts.Classifier](k, m) }
}

// WithBoundaries registers additional boundaries (e.g., wire.RegisterBoundaries).
func WithBoundaries(fn func(k sum.Key)) RegistryOption {
	return func(k sum.Key) { fn(k) }
}

// SetupRegistry creates a fresh sum registry with the given mock contracts,
// model boundaries, and wire boundaries. Returns context.Background().
func SetupRegistry(t *testing.T, opts ...RegistryOption) context.Context {
	t.Helper()
	sum.Reset()
	sum.New()
	k := sum.Start()

	// Load configs with defaults for gRPC timeout support.
	bgCtx := context.Background()
	_ = sum.Config[config.OCR](bgCtx, k, nil)
	_ = sum.Config[config.Convert](bgCtx, k, nil)
	_ = sum.Config[config.Classify](bgCtx, k, nil)

	for _, opt := range opts {
		opt(k)
	}

	// Model boundaries.
	sum.NewBoundary[models.Tenant](k)
	sum.NewBoundary[models.Provider](k)
	sum.NewBoundary[models.WatchedPath](k)
	sum.NewBoundary[models.Document](k)
	sum.NewBoundary[models.DocumentVersion](k)
	sum.NewBoundary[models.User](k)
	sum.NewBoundary[models.Subscription](k)

	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	return context.Background()
}

// SetupAPIEngine creates a rocco test engine with authenticated mock identity
// (tenant-1) and all API handlers registered.
func SetupAPIEngine(t *testing.T, handlers []rocco.Endpoint, opts ...RegistryOption) *rocco.Engine {
	t.Helper()
	_ = SetupRegistry(t, opts...)

	identity := rtesting.NewMockIdentity("user-1").WithTenantID("tenant-1")
	return rtesting.TestEngineWithAuth(func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return identity, nil
	}).WithHandlers(handlers...)
}

// SetupAdminEngine creates a rocco test engine with admin mock identity
// and all admin handlers registered.
func SetupAdminEngine(t *testing.T, handlers []rocco.Endpoint, opts ...RegistryOption) *rocco.Engine {
	t.Helper()
	_ = SetupRegistry(t, opts...)

	identity := rtesting.NewMockIdentity("admin-1").WithRoles("admin")
	return rtesting.TestEngineWithAuth(func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return identity, nil
	}).WithHandlers(handlers...)
}
