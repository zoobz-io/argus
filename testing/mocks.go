//go:build testing

package argustest

import (
	"context"

	admincontracts "github.com/zoobz-io/argus/admin/contracts"
	apicontracts "github.com/zoobz-io/argus/api/contracts"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/proto"
	"github.com/zoobz-io/vex"
	"google.golang.org/grpc"
)

// Compile-time interface assertions — prevents mock drift.
var (
	_ admincontracts.Tenants              = (*MockTenants)(nil)
	_ apicontracts.Tenants                = (*MockTenants)(nil)
	_ admincontracts.Providers            = (*MockProviders)(nil)
	_ apicontracts.Providers              = (*MockProviders)(nil)
	_ admincontracts.WatchedPaths         = (*MockWatchedPaths)(nil)
	_ apicontracts.WatchedPaths           = (*MockWatchedPaths)(nil)
	_ admincontracts.Documents            = (*MockDocuments)(nil)
	_ apicontracts.Documents              = (*MockDocuments)(nil)
	_ admincontracts.DocumentVersions     = (*MockDocumentVersions)(nil)
	_ apicontracts.DocumentVersions       = (*MockDocumentVersions)(nil)
	_ admincontracts.DocumentVersionSearch = (*MockDocumentVersionSearch)(nil)
	_ apicontracts.DocumentVersionSearch  = (*MockDocumentVersionSearch)(nil)
	_ admincontracts.Topics               = (*MockTopics)(nil)
	_ apicontracts.Topics                 = (*MockTopics)(nil)
	_ admincontracts.Tags                 = (*MockTags)(nil)
	_ apicontracts.Tags                   = (*MockTags)(nil)
	_ apicontracts.Ingest                 = (*MockIngest)(nil)
	_ apicontracts.IngestEnqueuer         = (*MockIngestEnqueuer)(nil)
	_ apicontracts.JobReader              = (*MockJobReader)(nil)
	_ apicontracts.QueryEmbedder          = (*MockQueryEmbedder)(nil)
	_ intcontracts.OCR                    = (*MockOCR)(nil)
)

// MockTenants satisfies both api/contracts.Tenants and admin/contracts.Tenants.
type MockTenants struct {
	OnGetTenant    func(ctx context.Context, id string) (*models.Tenant, error)
	OnCreateTenant func(ctx context.Context, name, slug string) (*models.Tenant, error)
	OnUpdateTenant func(ctx context.Context, id, name, slug string) (*models.Tenant, error)
	OnDeleteTenant func(ctx context.Context, id string) error
	OnListTenants  func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Tenant], error)
}

func (m *MockTenants) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	if m.OnGetTenant != nil { return m.OnGetTenant(ctx, id) }
	return &models.Tenant{}, nil
}
func (m *MockTenants) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	if m.OnCreateTenant != nil { return m.OnCreateTenant(ctx, name, slug) }
	return &models.Tenant{}, nil
}
func (m *MockTenants) UpdateTenant(ctx context.Context, id, name, slug string) (*models.Tenant, error) {
	if m.OnUpdateTenant != nil { return m.OnUpdateTenant(ctx, id, name, slug) }
	return &models.Tenant{}, nil
}
func (m *MockTenants) DeleteTenant(ctx context.Context, id string) error {
	if m.OnDeleteTenant != nil { return m.OnDeleteTenant(ctx, id) }
	return nil
}
func (m *MockTenants) ListTenants(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Tenant], error) {
	if m.OnListTenants != nil { return m.OnListTenants(ctx, page) }
	return &models.OffsetResult[models.Tenant]{Items: []*models.Tenant{}}, nil
}

// MockProviders satisfies both api/contracts.Providers and admin/contracts.Providers.
type MockProviders struct {
	OnGetProvider           func(ctx context.Context, id string) (*models.Provider, error)
	OnCreateProvider        func(ctx context.Context, tenantID string, pt models.ProviderType, name, creds string) (*models.Provider, error)
	OnUpdateProvider        func(ctx context.Context, id string, pt models.ProviderType, name, creds string) (*models.Provider, error)
	OnDeleteProvider        func(ctx context.Context, id string) error
	OnListProviders         func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
	OnListProvidersByTenant func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
}

func (m *MockProviders) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	if m.OnGetProvider != nil { return m.OnGetProvider(ctx, id) }
	return &models.Provider{}, nil
}
func (m *MockProviders) CreateProvider(ctx context.Context, tenantID string, pt models.ProviderType, name, creds string) (*models.Provider, error) {
	if m.OnCreateProvider != nil { return m.OnCreateProvider(ctx, tenantID, pt, name, creds) }
	return &models.Provider{}, nil
}
func (m *MockProviders) UpdateProvider(ctx context.Context, id string, pt models.ProviderType, name, creds string) (*models.Provider, error) {
	if m.OnUpdateProvider != nil { return m.OnUpdateProvider(ctx, id, pt, name, creds) }
	return &models.Provider{}, nil
}
func (m *MockProviders) DeleteProvider(ctx context.Context, id string) error {
	if m.OnDeleteProvider != nil { return m.OnDeleteProvider(ctx, id) }
	return nil
}
func (m *MockProviders) ListProviders(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
	if m.OnListProviders != nil { return m.OnListProviders(ctx, page) }
	return &models.OffsetResult[models.Provider]{Items: []*models.Provider{}}, nil
}
func (m *MockProviders) ListProvidersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
	if m.OnListProvidersByTenant != nil { return m.OnListProvidersByTenant(ctx, tenantID, page) }
	return &models.OffsetResult[models.Provider]{Items: []*models.Provider{}}, nil
}

// MockWatchedPaths satisfies both api/contracts.WatchedPaths and admin/contracts.WatchedPaths.
type MockWatchedPaths struct {
	OnGetWatchedPath           func(ctx context.Context, id string) (*models.WatchedPath, error)
	OnCreateWatchedPath        func(ctx context.Context, tenantID, providerID, path string) (*models.WatchedPath, error)
	OnUpdateWatchedPath        func(ctx context.Context, id, path string) (*models.WatchedPath, error)
	OnDeleteWatchedPath        func(ctx context.Context, id string) error
	OnListWatchedPaths         func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error)
	OnListWatchedPathsByTenant func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error)
}

func (m *MockWatchedPaths) GetWatchedPath(ctx context.Context, id string) (*models.WatchedPath, error) {
	if m.OnGetWatchedPath != nil { return m.OnGetWatchedPath(ctx, id) }
	return &models.WatchedPath{}, nil
}
func (m *MockWatchedPaths) CreateWatchedPath(ctx context.Context, tenantID, providerID, path string) (*models.WatchedPath, error) {
	if m.OnCreateWatchedPath != nil { return m.OnCreateWatchedPath(ctx, tenantID, providerID, path) }
	return &models.WatchedPath{}, nil
}
func (m *MockWatchedPaths) UpdateWatchedPath(ctx context.Context, id, path string) (*models.WatchedPath, error) {
	if m.OnUpdateWatchedPath != nil { return m.OnUpdateWatchedPath(ctx, id, path) }
	return &models.WatchedPath{}, nil
}
func (m *MockWatchedPaths) DeleteWatchedPath(ctx context.Context, id string) error {
	if m.OnDeleteWatchedPath != nil { return m.OnDeleteWatchedPath(ctx, id) }
	return nil
}
func (m *MockWatchedPaths) ListWatchedPaths(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
	if m.OnListWatchedPaths != nil { return m.OnListWatchedPaths(ctx, page) }
	return &models.OffsetResult[models.WatchedPath]{Items: []*models.WatchedPath{}}, nil
}
func (m *MockWatchedPaths) ListWatchedPathsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
	if m.OnListWatchedPathsByTenant != nil { return m.OnListWatchedPathsByTenant(ctx, tenantID, page) }
	return &models.OffsetResult[models.WatchedPath]{Items: []*models.WatchedPath{}}, nil
}

// MockDocuments satisfies both api/contracts.Documents and admin/contracts.Documents.
type MockDocuments struct {
	OnGetDocument          func(ctx context.Context, id string) (*models.Document, error)
	OnDeleteDocument       func(ctx context.Context, id string) error
	OnListDocuments        func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Document], error)
	OnListDocumentsByTenant func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Document], error)
}

func (m *MockDocuments) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	if m.OnGetDocument != nil { return m.OnGetDocument(ctx, id) }
	return &models.Document{}, nil
}
func (m *MockDocuments) DeleteDocument(ctx context.Context, id string) error {
	if m.OnDeleteDocument != nil { return m.OnDeleteDocument(ctx, id) }
	return nil
}
func (m *MockDocuments) ListDocuments(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Document], error) {
	if m.OnListDocuments != nil { return m.OnListDocuments(ctx, page) }
	return &models.OffsetResult[models.Document]{Items: []*models.Document{}}, nil
}
func (m *MockDocuments) ListDocumentsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Document], error) {
	if m.OnListDocumentsByTenant != nil { return m.OnListDocumentsByTenant(ctx, tenantID, page) }
	return &models.OffsetResult[models.Document]{Items: []*models.Document{}}, nil
}

// MockDocumentVersions satisfies both api/contracts.DocumentVersions and admin/contracts.DocumentVersions.
type MockDocumentVersions struct {
	OnGetDocumentVersion    func(ctx context.Context, id string) (*models.DocumentVersion, error)
	OnDeleteDocumentVersion func(ctx context.Context, id string) error
	OnListDocumentVersions  func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error)
	OnListVersionsByDocument func(ctx context.Context, docID string, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error)
}

func (m *MockDocumentVersions) GetDocumentVersion(ctx context.Context, id string) (*models.DocumentVersion, error) {
	if m.OnGetDocumentVersion != nil { return m.OnGetDocumentVersion(ctx, id) }
	return &models.DocumentVersion{}, nil
}
func (m *MockDocumentVersions) DeleteDocumentVersion(ctx context.Context, id string) error {
	if m.OnDeleteDocumentVersion != nil { return m.OnDeleteDocumentVersion(ctx, id) }
	return nil
}
func (m *MockDocumentVersions) ListDocumentVersions(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
	if m.OnListDocumentVersions != nil { return m.OnListDocumentVersions(ctx, page) }
	return &models.OffsetResult[models.DocumentVersion]{Items: []*models.DocumentVersion{}}, nil
}
func (m *MockDocumentVersions) ListVersionsByDocument(ctx context.Context, docID string, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
	if m.OnListVersionsByDocument != nil { return m.OnListVersionsByDocument(ctx, docID, page) }
	return &models.OffsetResult[models.DocumentVersion]{Items: []*models.DocumentVersion{}}, nil
}

// MockDocumentVersionSearch satisfies both api/contracts.DocumentVersionSearch and admin/contracts.DocumentVersionSearch.
type MockDocumentVersionSearch struct {
	OnSearch               func(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error)
	OnGetDocumentEmbedding func(ctx context.Context, documentID string) ([]float32, error)
	OnIndexVersion         func(ctx context.Context, version *models.DocumentVersionIndex) error
	OnDeleteDocument       func(ctx context.Context, documentID string) error
}

func (m *MockDocumentVersionSearch) Search(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error) {
	if m.OnSearch != nil { return m.OnSearch(ctx, params) }
	return &models.SearchResult{}, nil
}
func (m *MockDocumentVersionSearch) GetDocumentEmbedding(ctx context.Context, documentID string) ([]float32, error) {
	if m.OnGetDocumentEmbedding != nil { return m.OnGetDocumentEmbedding(ctx, documentID) }
	return []float32{}, nil
}
func (m *MockDocumentVersionSearch) IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error {
	if m.OnIndexVersion != nil { return m.OnIndexVersion(ctx, version) }
	return nil
}
func (m *MockDocumentVersionSearch) DeleteDocument(ctx context.Context, documentID string) error {
	if m.OnDeleteDocument != nil { return m.OnDeleteDocument(ctx, documentID) }
	return nil
}

// MockTopics satisfies both api/contracts.Topics and admin/contracts.Topics.
type MockTopics struct {
	OnGetTopic          func(ctx context.Context, id string) (*models.Topic, error)
	OnCreateTopic       func(ctx context.Context, tenantID, name, desc string) (*models.Topic, error)
	OnUpdateTopic       func(ctx context.Context, id, name, desc string) (*models.Topic, error)
	OnDeleteTopic       func(ctx context.Context, id string) error
	OnListTopicsByTenant func(ctx context.Context, tenantID string) ([]*models.Topic, error)
}

func (m *MockTopics) GetTopic(ctx context.Context, id string) (*models.Topic, error) {
	if m.OnGetTopic != nil { return m.OnGetTopic(ctx, id) }
	return &models.Topic{}, nil
}
func (m *MockTopics) CreateTopic(ctx context.Context, tenantID, name, desc string) (*models.Topic, error) {
	if m.OnCreateTopic != nil { return m.OnCreateTopic(ctx, tenantID, name, desc) }
	return &models.Topic{}, nil
}
func (m *MockTopics) UpdateTopic(ctx context.Context, id, name, desc string) (*models.Topic, error) {
	if m.OnUpdateTopic != nil { return m.OnUpdateTopic(ctx, id, name, desc) }
	return &models.Topic{}, nil
}
func (m *MockTopics) DeleteTopic(ctx context.Context, id string) error {
	if m.OnDeleteTopic != nil { return m.OnDeleteTopic(ctx, id) }
	return nil
}
func (m *MockTopics) ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error) {
	if m.OnListTopicsByTenant != nil { return m.OnListTopicsByTenant(ctx, tenantID) }
	return []*models.Topic{}, nil
}

// MockTags satisfies both api/contracts.Tags and admin/contracts.Tags.
type MockTags struct {
	OnGetTag          func(ctx context.Context, id string) (*models.Tag, error)
	OnCreateTag       func(ctx context.Context, tenantID, name, desc string) (*models.Tag, error)
	OnUpdateTag       func(ctx context.Context, id, name, desc string) (*models.Tag, error)
	OnDeleteTag       func(ctx context.Context, id string) error
	OnListTagsByTenant func(ctx context.Context, tenantID string) ([]*models.Tag, error)
}

func (m *MockTags) GetTag(ctx context.Context, id string) (*models.Tag, error) {
	if m.OnGetTag != nil { return m.OnGetTag(ctx, id) }
	return &models.Tag{}, nil
}
func (m *MockTags) CreateTag(ctx context.Context, tenantID, name, desc string) (*models.Tag, error) {
	if m.OnCreateTag != nil { return m.OnCreateTag(ctx, tenantID, name, desc) }
	return &models.Tag{}, nil
}
func (m *MockTags) UpdateTag(ctx context.Context, id, name, desc string) (*models.Tag, error) {
	if m.OnUpdateTag != nil { return m.OnUpdateTag(ctx, id, name, desc) }
	return &models.Tag{}, nil
}
func (m *MockTags) DeleteTag(ctx context.Context, id string) error {
	if m.OnDeleteTag != nil { return m.OnDeleteTag(ctx, id) }
	return nil
}
func (m *MockTags) ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error) {
	if m.OnListTagsByTenant != nil { return m.OnListTagsByTenant(ctx, tenantID) }
	return []*models.Tag{}, nil
}

// MockIngest satisfies api/contracts.Ingest.
type MockIngest struct {
	OnIngest func(ctx context.Context, jobID, versionID string) error
}

func (m *MockIngest) Ingest(ctx context.Context, jobID, versionID string) error {
	if m.OnIngest != nil { return m.OnIngest(ctx, jobID, versionID) }
	return nil
}

// MockIngestEnqueuer satisfies api/contracts.IngestEnqueuer.
type MockIngestEnqueuer struct {
	OnEnqueue func(ctx context.Context, versionID string) (*models.Job, error)
}

func (m *MockIngestEnqueuer) Enqueue(ctx context.Context, versionID string) (*models.Job, error) {
	if m.OnEnqueue != nil { return m.OnEnqueue(ctx, versionID) }
	return &models.Job{ID: "mock-job", Status: models.JobPending}, nil
}

// MockJobReader satisfies api/contracts.JobReader.
type MockJobReader struct {
	OnGetJob func(ctx context.Context, id string) (*models.Job, error)
}

func (m *MockJobReader) GetJob(ctx context.Context, id string) (*models.Job, error) {
	if m.OnGetJob != nil { return m.OnGetJob(ctx, id) }
	return &models.Job{ID: id, Status: models.JobPending}, nil
}

// MockQueryEmbedder satisfies api/contracts.QueryEmbedder.
type MockQueryEmbedder struct {
	OnEmbedQuery func(ctx context.Context, text string) (vex.Vector, error)
}

func (m *MockQueryEmbedder) EmbedQuery(ctx context.Context, text string) (vex.Vector, error) {
	if m.OnEmbedQuery != nil { return m.OnEmbedQuery(ctx, text) }
	return vex.Vector{}, nil
}

// MockOCR satisfies internal/contracts.OCR (proto.OCRServiceClient).
type MockOCR struct {
	OnExtractText func(ctx context.Context, in *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error)
}

func (m *MockOCR) ExtractText(ctx context.Context, in *proto.ExtractTextRequest, _ ...grpc.CallOption) (*proto.ExtractTextResponse, error) {
	if m.OnExtractText != nil { return m.OnExtractText(ctx, in) }
	return &proto.ExtractTextResponse{}, nil
}

// MockConverter satisfies internal/contracts.Converter (proto.ConvertServiceClient).
type MockConverter struct {
	OnConvertDocument func(ctx context.Context, in *proto.ConvertRequest) (*proto.ConvertResponse, error)
}

func (m *MockConverter) ConvertDocument(ctx context.Context, in *proto.ConvertRequest, _ ...grpc.CallOption) (*proto.ConvertResponse, error) {
	if m.OnConvertDocument != nil { return m.OnConvertDocument(ctx, in) }
	return &proto.ConvertResponse{}, nil
}

// MockClassifier satisfies internal/contracts.Classifier (proto.ClassifyServiceClient).
type MockClassifier struct {
	OnClassifyText func(ctx context.Context, in *proto.ClassifyRequest) (*proto.ClassifyResponse, error)
}

func (m *MockClassifier) ClassifyText(ctx context.Context, in *proto.ClassifyRequest, _ ...grpc.CallOption) (*proto.ClassifyResponse, error) {
	if m.OnClassifyText != nil { return m.OnClassifyText(ctx, in) }
	return &proto.ClassifyResponse{Safe: true}, nil
}

// MockVocabulary satisfies both api/contracts.Vocabulary and admin/contracts.Vocabulary.
type MockVocabulary struct {
	OnProcess       func(ctx context.Context, tenantID, name, description string) error
	OnProcessUpdate func(ctx context.Context, id, name, description string) error
}

func (m *MockVocabulary) Process(ctx context.Context, tenantID, name, description string) error {
	if m.OnProcess != nil { return m.OnProcess(ctx, tenantID, name, description) }
	return nil
}
func (m *MockVocabulary) ProcessUpdate(ctx context.Context, id, name, description string) error {
	if m.OnProcessUpdate != nil { return m.OnProcessUpdate(ctx, id, name, description) }
	return nil
}
