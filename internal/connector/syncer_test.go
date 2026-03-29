package connector

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// mockSyncStore implements SyncStore for testing.
type mockSyncStore struct {
	onListActiveWatchedPaths func(ctx context.Context) ([]*models.WatchedPath, error)
	onUpdateSyncState        func(ctx context.Context, id string, syncState *string) error
	onGetDocumentByExternalID func(ctx context.Context, tenantID, externalID string) (*models.Document, error)
	onCreateDocument         func(ctx context.Context, doc *models.Document) (*models.Document, error)
	onCreateDocumentVersion  func(ctx context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error)
	onGetLatestVersion       func(ctx context.Context, documentID string) (*models.DocumentVersion, error)
	onGetProvider            func(ctx context.Context, id string) (*models.Provider, error)
}

func (m *mockSyncStore) ListActiveWatchedPaths(ctx context.Context) ([]*models.WatchedPath, error) {
	if m.onListActiveWatchedPaths != nil {
		return m.onListActiveWatchedPaths(ctx)
	}
	return nil, nil
}

func (m *mockSyncStore) UpdateSyncState(ctx context.Context, id string, syncState *string) error {
	if m.onUpdateSyncState != nil {
		return m.onUpdateSyncState(ctx, id, syncState)
	}
	return nil
}

func (m *mockSyncStore) GetDocumentByExternalID(ctx context.Context, tenantID, externalID string) (*models.Document, error) {
	if m.onGetDocumentByExternalID != nil {
		return m.onGetDocumentByExternalID(ctx, tenantID, externalID)
	}
	return nil, nil
}

func (m *mockSyncStore) CreateDocument(ctx context.Context, doc *models.Document) (*models.Document, error) {
	if m.onCreateDocument != nil {
		return m.onCreateDocument(ctx, doc)
	}
	return doc, nil
}

func (m *mockSyncStore) CreateDocumentVersion(ctx context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error) {
	if m.onCreateDocumentVersion != nil {
		return m.onCreateDocumentVersion(ctx, ver)
	}
	return ver, nil
}

func (m *mockSyncStore) GetLatestVersion(ctx context.Context, documentID string) (*models.DocumentVersion, error) {
	if m.onGetLatestVersion != nil {
		return m.onGetLatestVersion(ctx, documentID)
	}
	return nil, nil
}

func (m *mockSyncStore) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	if m.onGetProvider != nil {
		return m.onGetProvider(ctx, id)
	}
	return &models.Provider{ID: id, Type: "test_provider"}, nil
}

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	providerType string
	onChanges    func(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error)
}

func (m *mockProvider) Type() string { return m.providerType }
func (m *mockProvider) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *mockProvider) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return nil, nil
}
func (m *mockProvider) List(_ context.Context, _ *provider.Credentials, _ string) ([]provider.Entry, *provider.Credentials, error) {
	return nil, nil, nil
}
func (m *mockProvider) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	if m.onChanges != nil {
		return m.onChanges(ctx, creds, path, syncToken)
	}
	return nil, "", nil, nil
}
func (m *mockProvider) Fetch(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	return nil, nil, nil, nil
}

// newTestSyncer builds a Syncer with the given mock store and provider.
func newTestSyncer(store SyncStore, prov provider.Provider, creds *provider.Credentials) *Syncer {
	registry := provider.NewRegistry()
	registry.Register(prov)

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: credJSON2(creds)}, nil
		},
	}
	credMgr := NewCredentialManager(provStore)

	return NewSyncer(store, credMgr, registry, time.Minute)
}

func credJSON2(creds *provider.Credentials) string {
	if creds == nil {
		return `{"access_token":"test-token"}`
	}
	return `{"access_token":"` + creds.AccessToken + `","refresh_token":"` + creds.RefreshToken + `"}`
}

func TestSyncer_PollPath_Success(t *testing.T) {
	var syncStateUpdated bool
	var docCreated bool
	var versionCreated bool

	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onUpdateSyncState: func(_ context.Context, _ string, state *string) error {
			syncStateUpdated = true
			if state == nil || *state != "token-2" {
				t.Errorf("expected sync state %q, got %v", "token-2", state)
			}
			return nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return nil, nil // Not found — will create.
		},
		onCreateDocument: func(_ context.Context, doc *models.Document) (*models.Document, error) {
			docCreated = true
			if doc.Name != "report.pdf" {
				t.Errorf("doc name: got %q, want %q", doc.Name, "report.pdf")
			}
			if doc.ExternalID != "file-ref-1" {
				t.Errorf("external ID: got %q, want %q", doc.ExternalID, "file-ref-1")
			}
			return doc, nil
		},
		onCreateDocumentVersion: func(_ context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error) {
			versionCreated = true
			if ver.ContentHash != "hash123" {
				t.Errorf("content hash: got %q, want %q", ver.ContentHash, "hash123")
			}
			return ver, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _ string, token string) ([]provider.Change, string, *provider.Credentials, error) {
			if token != syncToken {
				t.Errorf("sync token: got %q, want %q", token, syncToken)
			}
			return []provider.Change{
				{
					Ref:  "file-ref-1",
					Type: provider.ChangeCreated,
					Entry: &provider.Entry{
						Ref:         "file-ref-1",
						Name:        "report.pdf",
						MimeType:    "application/pdf",
						ContentHash: "hash123",
						Size:        1024,
						ModifiedAt:  time.Now(),
					},
				},
			}, "token-2", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		Active:     true,
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !syncStateUpdated {
		t.Error("sync state was not updated")
	}
	if !docCreated {
		t.Error("document was not created")
	}
	if !versionCreated {
		t.Error("document version was not created")
	}
}

func TestSyncer_PollPath_InitialSync(t *testing.T) {
	var syncStateUpdated bool

	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onUpdateSyncState: func(_ context.Context, _ string, state *string) error {
			syncStateUpdated = true
			if state == nil || *state != "initial-token" {
				t.Errorf("expected sync state %q, got %v", "initial-token", state)
			}
			return nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return nil, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _ string, token string) ([]provider.Change, string, *provider.Credentials, error) {
			if token != "" {
				t.Errorf("expected empty sync token for initial sync, got %q", token)
			}
			return nil, "initial-token", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		Active:     true,
		SyncState:  nil, // No previous sync.
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !syncStateUpdated {
		t.Error("sync state was not updated with initial token")
	}
}

func TestSyncer_PollPath_NoChanges(t *testing.T) {
	var docCreated bool

	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateDocument: func(_ context.Context, _ *models.Document) (*models.Document, error) {
			docCreated = true
			return nil, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _ string, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return nil, "token-1", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		Active:     true,
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docCreated {
		t.Error("should not create documents when there are no changes")
	}
}

func TestSyncer_PollPath_CredentialRefresh(t *testing.T) {
	var credsUpdated bool

	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	refreshedCreds := &provider.Credentials{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: `{"access_token":"old-token","refresh_token":"old-refresh"}`}, nil
		},
		onUpdateProviderCredentials: func(_ context.Context, _, _, _ string) error {
			credsUpdated = true
			return nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _ string, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return nil, "token-2", refreshedCreds, nil
		},
	}

	registry := provider.NewRegistry()
	registry.Register(prov)
	credMgr := NewCredentialManager(provStore)

	syncer := NewSyncer(store, credMgr, registry, time.Minute)

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		Active:     true,
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !credsUpdated {
		t.Error("credentials were not persisted after refresh")
	}
}

func TestSyncer_PollPath_ProviderError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _ string, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return nil, "", nil, errors.New("api rate limited")
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		Active:     true,
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	// The error should be returned but pollAll logs it and continues — the syncer does not crash.
}

func TestSyncer_Run_ShutdownOnCancel(t *testing.T) {
	store := &mockSyncStore{
		onListActiveWatchedPaths: func(_ context.Context) ([]*models.WatchedPath, error) {
			return nil, nil
		},
	}

	prov := &mockProvider{providerType: "test_provider"}
	syncer := newTestSyncer(store, prov, nil)
	syncer.interval = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- syncer.Run(ctx)
	}()

	// Let it tick once then cancel.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("syncer did not shut down in time")
	}
}

func TestSyncer_PollPath_DedupByContentHash(t *testing.T) {
	var versionCreated bool

	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return &models.Document{ID: "doc-1", ExternalID: "ref-1"}, nil
		},
		onGetLatestVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: "ver-1", ContentHash: "same-hash"}, nil
		},
		onCreateDocumentVersion: func(_ context.Context, _ *models.DocumentVersion) (*models.DocumentVersion, error) {
			versionCreated = true
			return nil, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "ref-1",
					Type: provider.ChangeModified,
					Entry: &provider.Entry{
						Ref:         "ref-1",
						Name:        "report.pdf",
						MimeType:    "application/pdf",
						ContentHash: "same-hash",
					},
				},
			}, "token-2", nil, nil
		},
	}

	reg := provider.NewRegistry()
	reg.Register(prov)

	credStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: `{"access_token":"token"}`}, nil
		},
	}
	cm := NewCredentialManager(credStore)

	wp := &models.WatchedPath{ID: "wp-1", ProviderID: "prov-1", TenantID: "tenant-1", Path: "/docs", SyncState: &syncToken}
	syncer := NewSyncer(store, cm, reg, time.Minute)

	if err := syncer.pollPath(context.Background(), wp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if versionCreated {
		t.Error("should not create version when content hash matches latest")
	}
}

func TestSyncer_PollAll_Success(t *testing.T) {
	var polledPaths []string

	syncToken := "token-1"
	store := &mockSyncStore{
		onListActiveWatchedPaths: func(_ context.Context) ([]*models.WatchedPath, error) {
			return []*models.WatchedPath{
				{ID: "wp-1", ProviderID: "prov-1", TenantID: "t-1", Path: "/docs", Active: true, SyncState: &syncToken},
				{ID: "wp-2", ProviderID: "prov-1", TenantID: "t-1", Path: "/reports", Active: true, SyncState: &syncToken},
			}, nil
		},
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, path string, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			polledPaths = append(polledPaths, path)
			return nil, "token-2", nil, nil
		},
	}

	reg := provider.NewRegistry()
	reg.Register(prov)

	credStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: `{"access_token":"token"}`}, nil
		},
	}
	cm := NewCredentialManager(credStore)

	syncer := NewSyncer(store, cm, reg, time.Minute)
	syncer.pollAll(context.Background())

	if len(polledPaths) != 2 {
		t.Fatalf("expected 2 paths polled, got %d", len(polledPaths))
	}
}

func TestSyncer_PollAll_ListError(t *testing.T) {
	store := &mockSyncStore{
		onListActiveWatchedPaths: func(_ context.Context) ([]*models.WatchedPath, error) {
			return nil, errors.New("db error")
		},
	}

	syncer := NewSyncer(store, nil, nil, time.Minute)
	// Should not panic — logs and returns.
	syncer.pollAll(context.Background())
}

func TestSyncer_PollAll_Empty(t *testing.T) {
	store := &mockSyncStore{
		onListActiveWatchedPaths: func(_ context.Context) ([]*models.WatchedPath, error) {
			return nil, nil
		},
	}

	syncer := NewSyncer(store, nil, nil, time.Minute)
	syncer.pollAll(context.Background())
}

func TestSyncer_PollPath_GetProviderError(t *testing.T) {
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, errors.New("provider not found")
		},
	}

	syncer := newTestSyncer(store, &mockProvider{providerType: "test_provider"}, nil)

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
	}

	err := syncer.pollPath(context.Background(), wp)
	if err == nil {
		t.Fatal("expected error from GetProvider")
	}
}

func TestSyncer_PollPath_RegistryLookupError(t *testing.T) {
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "unknown_provider"}, nil
		},
	}

	// Registry has "test_provider" but store returns "unknown_provider".
	syncer := newTestSyncer(store, &mockProvider{providerType: "test_provider"}, nil)

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
	}

	err := syncer.pollPath(context.Background(), wp)
	if err == nil {
		t.Fatal("expected error from registry lookup")
	}
}

func TestSyncer_PollPath_UpdateSyncStateError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onUpdateSyncState: func(_ context.Context, _ string, _ *string) error {
			return errors.New("db write error")
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return nil, "new-token", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err == nil {
		t.Fatal("expected error from UpdateSyncState")
	}
}

func TestSyncer_PollPath_CredentialUpdateError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	refreshedCreds := &provider.Credentials{AccessToken: "new-token"}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return nil, "", refreshedCreds, nil
		},
	}

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: `{"access_token":"old"}`}, nil
		},
		onUpdateProviderCredentials: func(_ context.Context, _, _, _ string) error {
			return errors.New("creds persist error")
		},
	}

	registry := provider.NewRegistry()
	registry.Register(prov)
	credMgr := NewCredentialManager(provStore)

	syncer := NewSyncer(store, credMgr, registry, time.Minute)

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err == nil {
		t.Fatal("expected error from credential update")
	}
}

func TestSyncer_PollPath_DeletedChangeSkipped(t *testing.T) {
	var versionCreated bool
	syncToken := "token-1"

	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateDocumentVersion: func(_ context.Context, _ *models.DocumentVersion) (*models.DocumentVersion, error) {
			versionCreated = true
			return nil, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "file-ref-1",
					Type: provider.ChangeDeleted,
				},
			}, "token-2", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if versionCreated {
		t.Error("should not create version for deleted change")
	}
}

func TestSyncer_PollPath_NilEntrySkipped(t *testing.T) {
	var versionCreated bool
	syncToken := "token-1"

	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateDocumentVersion: func(_ context.Context, _ *models.DocumentVersion) (*models.DocumentVersion, error) {
			versionCreated = true
			return nil, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:   "file-ref-1",
					Type:  provider.ChangeCreated,
					Entry: nil, // nil entry
				},
			}, "token-2", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if versionCreated {
		t.Error("should not create version for nil entry")
	}
}

func TestSyncer_ProcessChange_GetDocumentByExternalIDError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return nil, errors.New("db lookup error")
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "file-ref-1",
					Type: provider.ChangeCreated,
					Entry: &provider.Entry{
						Ref:      "file-ref-1",
						Name:     "report.pdf",
						MimeType: "application/pdf",
					},
				},
			}, "", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	// pollPath logs processChange errors but does not return them.
	// Verify it does not panic.
	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncer_ProcessChange_CreateDocumentError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return nil, nil // Not found.
		},
		onCreateDocument: func(_ context.Context, _ *models.Document) (*models.Document, error) {
			return nil, errors.New("db create error")
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "file-ref-1",
					Type: provider.ChangeCreated,
					Entry: &provider.Entry{
						Ref:      "file-ref-1",
						Name:     "report.pdf",
						MimeType: "application/pdf",
					},
				},
			}, "", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncer_ProcessChange_CreateDocumentVersionError(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return nil, nil // Not found.
		},
		onCreateDocument: func(_ context.Context, doc *models.Document) (*models.Document, error) {
			return doc, nil
		},
		onCreateDocumentVersion: func(_ context.Context, _ *models.DocumentVersion) (*models.DocumentVersion, error) {
			return nil, errors.New("db version error")
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "file-ref-1",
					Type: provider.ChangeCreated,
					Entry: &provider.Entry{
						Ref:         "file-ref-1",
						Name:        "report.pdf",
						MimeType:    "application/pdf",
						ContentHash: "hash-new",
					},
				},
			}, "", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncer_ProcessChange_ExistingDocNewVersion(t *testing.T) {
	var versionCreated bool
	syncToken := "token-1"

	store := &mockSyncStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onGetDocumentByExternalID: func(_ context.Context, _, _ string) (*models.Document, error) {
			return &models.Document{ID: "doc-1", ExternalID: "ref-1", ObjectKey: "objects/t-1/doc-1"}, nil
		},
		onGetLatestVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: "ver-1", ContentHash: "old-hash"}, nil
		},
		onCreateDocumentVersion: func(_ context.Context, ver *models.DocumentVersion) (*models.DocumentVersion, error) {
			versionCreated = true
			return ver, nil
		},
	}

	prov := &mockProvider{
		providerType: "test_provider",
		onChanges: func(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
			return []provider.Change{
				{
					Ref:  "ref-1",
					Type: provider.ChangeModified,
					Entry: &provider.Entry{
						Ref:         "ref-1",
						Name:        "report.pdf",
						MimeType:    "application/pdf",
						ContentHash: "new-hash",
					},
				},
			}, "token-2", nil, nil
		},
	}

	syncer := newTestSyncer(store, prov, &provider.Credentials{AccessToken: "test-token"})

	wp := &models.WatchedPath{
		ID:         "wp-1",
		TenantID:   "t-1",
		ProviderID: "p-1",
		Path:       "/docs",
		SyncState:  &syncToken,
	}

	err := syncer.pollPath(context.Background(), wp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !versionCreated {
		t.Error("expected new version to be created for changed content hash")
	}
}

func TestSyncer_PollAll_PollPathError_DoesNotCrash(t *testing.T) {
	syncToken := "token-1"
	store := &mockSyncStore{
		onListActiveWatchedPaths: func(_ context.Context) ([]*models.WatchedPath, error) {
			return []*models.WatchedPath{
				{ID: "wp-1", ProviderID: "prov-1", TenantID: "t-1", Path: "/docs", SyncState: &syncToken},
			}, nil
		},
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, errors.New("provider error")
		},
	}

	syncer := newTestSyncer(store, &mockProvider{providerType: "test_provider"}, nil)
	// Should not panic — logs the error and continues.
	syncer.pollAll(context.Background())
}
