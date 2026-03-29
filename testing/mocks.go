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
	_ apicontracts.AuditLog               = (*MockAuditLog)(nil)
	_ admincontracts.AuditLog             = (*MockAdminAuditLog)(nil)
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
	_ admincontracts.Users                = (*MockUsers)(nil)
	_ apicontracts.Users                  = (*MockUsers)(nil)
	_ apicontracts.Ingest                 = (*MockIngest)(nil)
	_ apicontracts.IngestEnqueuer         = (*MockIngestEnqueuer)(nil)
	_ apicontracts.JobReader              = (*MockJobReader)(nil)
	_ apicontracts.QueryEmbedder          = (*MockQueryEmbedder)(nil)
	_ intcontracts.OCR                    = (*MockOCR)(nil)
	_ admincontracts.Subscriptions        = (*MockAdminSubscriptions)(nil)
	_ apicontracts.Subscriptions          = (*MockSubscriptions)(nil)
	_ apicontracts.Notifications          = (*MockNotifications)(nil)
	_ apicontracts.Hooks                  = (*MockHooks)(nil)
	_ apicontracts.Deliveries             = (*MockDeliveries)(nil)
	_ admincontracts.Hooks                = (*MockAdminHooks)(nil)
	_ intcontracts.NotifyHookLoader       = (*MockHookLoader)(nil)
	_ intcontracts.NotifyDeliveryLogger   = (*MockDeliveryLogger)(nil)
)

// MockAuditLog satisfies api/contracts.AuditLog.
type MockAuditLog struct {
	OnSearch func(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error)
}

func (m *MockAuditLog) Search(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
	if m.OnSearch != nil { return m.OnSearch(ctx, params) }
	return &models.OffsetResult[models.AuditEntry]{Items: []*models.AuditEntry{}}, nil
}

// MockAdminAuditLog satisfies admin/contracts.AuditLog.
type MockAdminAuditLog struct {
	OnSearch func(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error)
}

func (m *MockAdminAuditLog) Search(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error) {
	if m.OnSearch != nil { return m.OnSearch(ctx, params) }
	return &models.OffsetResult[models.AuditEntry]{Items: []*models.AuditEntry{}}, nil
}

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
	OnGetProvider              func(ctx context.Context, id string) (*models.Provider, error)
	OnGetProviderByTenant      func(ctx context.Context, id, tenantID string) (*models.Provider, error)
	OnCreateProvider           func(ctx context.Context, tenantID string, pt models.ProviderType, name, creds string) (*models.Provider, error)
	OnUpdateProvider           func(ctx context.Context, id string, pt models.ProviderType, name, creds string) (*models.Provider, error)
	OnUpdateProviderCredentials func(ctx context.Context, id, credentials string) error
	OnDeleteProvider           func(ctx context.Context, id string) error
	OnListProviders            func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
	OnListProvidersByTenant    func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
}

func (m *MockProviders) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	if m.OnGetProvider != nil { return m.OnGetProvider(ctx, id) }
	return &models.Provider{}, nil
}
func (m *MockProviders) GetProviderByTenant(ctx context.Context, id, tenantID string) (*models.Provider, error) {
	if m.OnGetProviderByTenant != nil { return m.OnGetProviderByTenant(ctx, id, tenantID) }
	return &models.Provider{ID: id, TenantID: tenantID}, nil
}
func (m *MockProviders) UpdateProviderCredentials(ctx context.Context, id, credentials string) error {
	if m.OnUpdateProviderCredentials != nil { return m.OnUpdateProviderCredentials(ctx, id, credentials) }
	return nil
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

// MockUsers satisfies both api/contracts.Users and admin/contracts.Users.
type MockUsers struct {
	OnGetUser            func(ctx context.Context, id string) (*models.User, error)
	OnGetUserByExternalID func(ctx context.Context, externalID string) (*models.User, error)
	OnCreateUser         func(ctx context.Context, tenantID, externalID, email, displayName string, role models.UserRole) (*models.User, error)
	OnUpdateUser         func(ctx context.Context, id, email, displayName string, role models.UserRole, status models.UserStatus) (*models.User, error)
	OnDeleteUser         func(ctx context.Context, id string) error
	OnListUsers          func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error)
	OnListUsersByTenant  func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error)
}

func (m *MockUsers) GetUser(ctx context.Context, id string) (*models.User, error) {
	if m.OnGetUser != nil { return m.OnGetUser(ctx, id) }
	return &models.User{}, nil
}
func (m *MockUsers) GetUserByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	if m.OnGetUserByExternalID != nil { return m.OnGetUserByExternalID(ctx, externalID) }
	return &models.User{}, nil
}
func (m *MockUsers) CreateUser(ctx context.Context, tenantID, externalID, email, displayName string, role models.UserRole) (*models.User, error) {
	if m.OnCreateUser != nil { return m.OnCreateUser(ctx, tenantID, externalID, email, displayName, role) }
	return &models.User{}, nil
}
func (m *MockUsers) UpdateUser(ctx context.Context, id, email, displayName string, role models.UserRole, status models.UserStatus) (*models.User, error) {
	if m.OnUpdateUser != nil { return m.OnUpdateUser(ctx, id, email, displayName, role, status) }
	return &models.User{}, nil
}
func (m *MockUsers) DeleteUser(ctx context.Context, id string) error {
	if m.OnDeleteUser != nil { return m.OnDeleteUser(ctx, id) }
	return nil
}
func (m *MockUsers) ListUsers(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	if m.OnListUsers != nil { return m.OnListUsers(ctx, page) }
	return &models.OffsetResult[models.User]{Items: []*models.User{}}, nil
}
func (m *MockUsers) ListUsersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	if m.OnListUsersByTenant != nil { return m.OnListUsersByTenant(ctx, tenantID, page) }
	return &models.OffsetResult[models.User]{Items: []*models.User{}}, nil
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
	OnEnqueue func(ctx context.Context, versionID, tenantID string) (*models.Job, error)
}

func (m *MockIngestEnqueuer) Enqueue(ctx context.Context, versionID, tenantID string) (*models.Job, error) {
	if m.OnEnqueue != nil { return m.OnEnqueue(ctx, versionID, tenantID) }
	return &models.Job{ID: "mock-job", Status: models.JobPending}, nil
}

// MockJobReader satisfies api/contracts.JobReader.
type MockJobReader struct {
	OnGetJobByTenant func(ctx context.Context, id, tenantID string) (*models.Job, error)
}

func (m *MockJobReader) GetJobByTenant(ctx context.Context, id, tenantID string) (*models.Job, error) {
	if m.OnGetJobByTenant != nil { return m.OnGetJobByTenant(ctx, id, tenantID) }
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

// MockSubscriptions satisfies api/contracts.Subscriptions.
type MockSubscriptions struct {
	OnGetSubscriptionByTenant func(ctx context.Context, tenantID, id string) (*models.Subscription, error)
	OnListSubscriptionsByUser func(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error)
	OnCreateSubscription      func(ctx context.Context, tenantID, userID, eventType string, channel models.SubscriptionChannel, webhookEndpointID string) (*models.Subscription, error)
	OnDeleteSubscription      func(ctx context.Context, tenantID, userID, id string) error
}

func (m *MockSubscriptions) GetSubscriptionByTenant(ctx context.Context, tenantID, id string) (*models.Subscription, error) {
	if m.OnGetSubscriptionByTenant != nil { return m.OnGetSubscriptionByTenant(ctx, tenantID, id) }
	return &models.Subscription{}, nil
}
func (m *MockSubscriptions) ListSubscriptionsByUser(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error) {
	if m.OnListSubscriptionsByUser != nil { return m.OnListSubscriptionsByUser(ctx, tenantID, userID, page) }
	return &models.OffsetResult[models.Subscription]{Items: []*models.Subscription{}}, nil
}
func (m *MockSubscriptions) CreateSubscription(ctx context.Context, tenantID, userID, eventType string, channel models.SubscriptionChannel, webhookEndpointID string) (*models.Subscription, error) {
	if m.OnCreateSubscription != nil { return m.OnCreateSubscription(ctx, tenantID, userID, eventType, channel, webhookEndpointID) }
	return &models.Subscription{}, nil
}
func (m *MockSubscriptions) DeleteSubscription(ctx context.Context, tenantID, userID, id string) error {
	if m.OnDeleteSubscription != nil { return m.OnDeleteSubscription(ctx, tenantID, userID, id) }
	return nil
}

// MockAdminSubscriptions satisfies admin/contracts.Subscriptions.
type MockAdminSubscriptions struct {
	OnGetSubscription    func(ctx context.Context, id string) (*models.Subscription, error)
	OnListSubscriptions  func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error)
	OnDeleteSubscription func(ctx context.Context, id string) error
}

func (m *MockAdminSubscriptions) GetSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	if m.OnGetSubscription != nil { return m.OnGetSubscription(ctx, id) }
	return &models.Subscription{}, nil
}
func (m *MockAdminSubscriptions) ListSubscriptions(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error) {
	if m.OnListSubscriptions != nil { return m.OnListSubscriptions(ctx, page) }
	return &models.OffsetResult[models.Subscription]{Items: []*models.Subscription{}}, nil
}
func (m *MockAdminSubscriptions) DeleteSubscription(ctx context.Context, id string) error {
	if m.OnDeleteSubscription != nil { return m.OnDeleteSubscription(ctx, id) }
	return nil
}

// MockNotifications satisfies api/contracts.Notifications.
type MockNotifications struct {
	OnSearchByUser    func(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Notification], error)
	OnUpdateStatus    func(ctx context.Context, tenantID, userID, id string, status models.NotificationStatus) (*models.Notification, error)
	OnBulkUpdateStatus func(ctx context.Context, tenantID, userID string, status models.NotificationStatus) error
}

func (m *MockNotifications) SearchByUser(ctx context.Context, tenantID, userID string, page models.OffsetPage) (*models.OffsetResult[models.Notification], error) {
	if m.OnSearchByUser != nil { return m.OnSearchByUser(ctx, tenantID, userID, page) }
	return &models.OffsetResult[models.Notification]{Items: []*models.Notification{}}, nil
}
func (m *MockNotifications) UpdateStatus(ctx context.Context, tenantID, userID, id string, status models.NotificationStatus) (*models.Notification, error) {
	if m.OnUpdateStatus != nil { return m.OnUpdateStatus(ctx, tenantID, userID, id, status) }
	return &models.Notification{}, nil
}
func (m *MockNotifications) BulkUpdateStatus(ctx context.Context, tenantID, userID string, status models.NotificationStatus) error {
	if m.OnBulkUpdateStatus != nil { return m.OnBulkUpdateStatus(ctx, tenantID, userID, status) }
	return nil
}

// MockHooks satisfies api/contracts.Hooks.
type MockHooks struct {
	OnCreateHook        func(ctx context.Context, tenantID, userID, url string) (*models.Hook, error)
	OnGetHookByTenant   func(ctx context.Context, tenantID, id string) (*models.Hook, error)
	OnListHooksByTenant func(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Hook], error)
	OnDeleteHook        func(ctx context.Context, tenantID, id string) error
}

func (m *MockHooks) CreateHook(ctx context.Context, tenantID, userID, url string) (*models.Hook, error) {
	if m.OnCreateHook != nil { return m.OnCreateHook(ctx, tenantID, userID, url) }
	return &models.Hook{}, nil
}
func (m *MockHooks) GetHookByTenant(ctx context.Context, tenantID, id string) (*models.Hook, error) {
	if m.OnGetHookByTenant != nil { return m.OnGetHookByTenant(ctx, tenantID, id) }
	return &models.Hook{}, nil
}
func (m *MockHooks) ListHooksByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
	if m.OnListHooksByTenant != nil { return m.OnListHooksByTenant(ctx, tenantID, page) }
	return &models.OffsetResult[models.Hook]{Items: []*models.Hook{}}, nil
}
func (m *MockHooks) DeleteHook(ctx context.Context, tenantID, id string) error {
	if m.OnDeleteHook != nil { return m.OnDeleteHook(ctx, tenantID, id) }
	return nil
}

// MockDeliveries satisfies api/contracts.Deliveries.
type MockDeliveries struct {
	OnListByHook func(ctx context.Context, tenantID, hookID string, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error)
}

func (m *MockDeliveries) ListByHook(ctx context.Context, tenantID, hookID string, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
	if m.OnListByHook != nil { return m.OnListByHook(ctx, tenantID, hookID, page) }
	return &models.OffsetResult[models.Delivery]{Items: []*models.Delivery{}}, nil
}

// MockAdminHooks satisfies admin/contracts.Hooks.
type MockAdminHooks struct {
	OnGetHook        func(ctx context.Context, id string) (*models.Hook, error)
	OnListHooks      func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Hook], error)
	OnDeleteHook     func(ctx context.Context, id string) error
	OnListDeliveries func(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error)
}

func (m *MockAdminHooks) GetHook(ctx context.Context, id string) (*models.Hook, error) {
	if m.OnGetHook != nil { return m.OnGetHook(ctx, id) }
	return &models.Hook{}, nil
}
func (m *MockAdminHooks) ListHooks(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
	if m.OnListHooks != nil { return m.OnListHooks(ctx, page) }
	return &models.OffsetResult[models.Hook]{Items: []*models.Hook{}}, nil
}
func (m *MockAdminHooks) DeleteHook(ctx context.Context, id string) error {
	if m.OnDeleteHook != nil { return m.OnDeleteHook(ctx, id) }
	return nil
}
func (m *MockAdminHooks) ListDeliveries(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
	if m.OnListDeliveries != nil { return m.OnListDeliveries(ctx, page) }
	return &models.OffsetResult[models.Delivery]{Items: []*models.Delivery{}}, nil
}

// MockHookLoader satisfies internal/contracts.NotifyHookLoader.
type MockHookLoader struct {
	OnGetWithSecret func(ctx context.Context, tenantID, id string) (*models.Hook, error)
}

func (m *MockHookLoader) GetWithSecret(ctx context.Context, tenantID, id string) (*models.Hook, error) {
	if m.OnGetWithSecret != nil { return m.OnGetWithSecret(ctx, tenantID, id) }
	return &models.Hook{}, nil
}

// MockDeliveryLogger satisfies internal/contracts.NotifyDeliveryLogger.
type MockDeliveryLogger struct {
	OnCreateDelivery func(ctx context.Context, hookID, eventID, tenantID string, statusCode, attempt int, deliveryErr *string) error
}

func (m *MockDeliveryLogger) CreateDelivery(ctx context.Context, hookID, eventID, tenantID string, statusCode, attempt int, deliveryErr *string) error {
	if m.OnCreateDelivery != nil { return m.OnCreateDelivery(ctx, hookID, eventID, tenantID, statusCode, attempt, deliveryErr) }
	return nil
}
