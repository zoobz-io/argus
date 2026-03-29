package connector

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/zoobz-io/grub"

	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// mockFetchStore implements FetchStore for testing.
type mockFetchStore struct {
	onGetProvider func(ctx context.Context, id string) (*models.Provider, error)
	onCreateJob   func(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error)
}

func (m *mockFetchStore) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	if m.onGetProvider != nil {
		return m.onGetProvider(ctx, id)
	}
	return &models.Provider{ID: id, Type: "test_provider"}, nil
}

func (m *mockFetchStore) CreateJob(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
	if m.onCreateJob != nil {
		return m.onCreateJob(ctx, versionID, documentID, tenantID)
	}
	return &models.Job{ID: "job-1", VersionID: versionID, DocumentID: documentID, TenantID: tenantID}, nil
}

// mockBucket implements grub.BucketProvider for testing.
type mockBucket struct {
	onPut    func(ctx context.Context, key string, data []byte, info *grub.ObjectInfo) error
	onGet    func(ctx context.Context, key string) ([]byte, *grub.ObjectInfo, error)
	onDelete func(ctx context.Context, key string) error
	onExists func(ctx context.Context, key string) (bool, error)
	onList   func(ctx context.Context, prefix string, limit int) ([]grub.ObjectInfo, error)
}

func (m *mockBucket) Put(ctx context.Context, key string, data []byte, info *grub.ObjectInfo) error {
	if m.onPut != nil {
		return m.onPut(ctx, key, data, info)
	}
	return nil
}

func (m *mockBucket) Get(ctx context.Context, key string) ([]byte, *grub.ObjectInfo, error) {
	if m.onGet != nil {
		return m.onGet(ctx, key)
	}
	return nil, nil, nil
}

func (m *mockBucket) Delete(ctx context.Context, key string) error {
	if m.onDelete != nil {
		return m.onDelete(ctx, key)
	}
	return nil
}

func (m *mockBucket) Exists(ctx context.Context, key string) (bool, error) {
	if m.onExists != nil {
		return m.onExists(ctx, key)
	}
	return false, nil
}

func (m *mockBucket) List(ctx context.Context, prefix string, limit int) ([]grub.ObjectInfo, error) {
	if m.onList != nil {
		return m.onList(ctx, prefix, limit)
	}
	return nil, nil
}

// mockFetchProvider implements provider.Provider with a configurable Fetch.
type mockFetchProvider struct {
	providerType string
	onFetch      func(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error)
}

func (m *mockFetchProvider) Type() string { return m.providerType }
func (m *mockFetchProvider) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *mockFetchProvider) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return nil, nil
}
func (m *mockFetchProvider) List(_ context.Context, _ *provider.Credentials, _ string) ([]provider.Entry, *provider.Credentials, error) {
	return nil, nil, nil
}
func (m *mockFetchProvider) Changes(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
	return nil, "", nil, nil
}
func (m *mockFetchProvider) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	if m.onFetch != nil {
		return m.onFetch(ctx, creds, ref)
	}
	return nil, nil, nil, nil
}

func newTestFetcher(store FetchStore, prov provider.Provider, creds *provider.Credentials, bucket grub.BucketProvider) *Fetcher {
	registry := provider.NewRegistry()
	registry.Register(prov)

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: credJSON2(creds)}, nil
		},
	}
	credMgr := NewCredentialManager(provStore)

	return NewFetcher(store, credMgr, registry, bucket)
}

func TestFetcher_HandleFetch_Success(t *testing.T) {
	content := []byte("file content here")
	var bucketKey string
	var bucketData []byte
	var jobCreated bool

	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateJob: func(_ context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
			jobCreated = true
			if versionID != "ver-1" {
				t.Errorf("version ID: got %q, want %q", versionID, "ver-1")
			}
			if documentID != "doc-1" {
				t.Errorf("document ID: got %q, want %q", documentID, "doc-1")
			}
			if tenantID != "t-1" {
				t.Errorf("tenant ID: got %q, want %q", tenantID, "t-1")
			}
			return &models.Job{ID: "job-1", VersionID: versionID, DocumentID: documentID, TenantID: tenantID}, nil
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			if ref != "file-ref-1" {
				t.Errorf("ref: got %q, want %q", ref, "file-ref-1")
			}
			return io.NopCloser(bytes.NewReader(content)), &provider.EntryMeta{
				Name:     "report.pdf",
				MimeType: "application/pdf",
				Size:     int64(len(content)),
			}, nil, nil
		},
	}

	bucket := &mockBucket{
		onPut: func(_ context.Context, key string, data []byte, info *grub.ObjectInfo) error {
			bucketKey = key
			bucketData = data
			if info.ContentType != "application/pdf" {
				t.Errorf("content type: got %q, want %q", info.ContentType, "application/pdf")
			}
			return nil
		},
	}

	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "file-ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	if err := fetcher.HandleFetch(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bucketKey != "objects/t-1/doc-1" {
		t.Errorf("bucket key: got %q, want %q", bucketKey, "objects/t-1/doc-1")
	}
	if !bytes.Equal(bucketData, content) {
		t.Errorf("bucket data mismatch: got %d bytes, want %d", len(bucketData), len(content))
	}
	if !jobCreated {
		t.Error("job was not created")
	}
}

func TestFetcher_HandleFetch_ProviderError(t *testing.T) {
	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			return nil, nil, nil, errors.New("rate limited")
		},
	}

	bucket := &mockBucket{}
	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "file-ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		// Just verify it wraps something meaningful.
		if err.Error() == "" {
			t.Error("error should have a message")
		}
	}
}

func TestFetcher_HandleFetch_CredentialRefresh(t *testing.T) {
	var credsUpdated bool
	content := []byte("data")

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

	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateJob: func(_ context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
			return &models.Job{ID: "job-1", VersionID: versionID, DocumentID: documentID, TenantID: tenantID}, nil
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			return io.NopCloser(bytes.NewReader(content)), &provider.EntryMeta{
				Name:     "file.txt",
				MimeType: "text/plain",
				Size:     int64(len(content)),
			}, refreshedCreds, nil
		},
	}

	bucket := &mockBucket{}

	registry := provider.NewRegistry()
	registry.Register(prov)
	credMgr := NewCredentialManager(provStore)

	fetcher := NewFetcher(store, credMgr, registry, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	if err := fetcher.HandleFetch(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !credsUpdated {
		t.Error("credentials were not persisted after refresh")
	}
}

func TestFetcher_HandleFetch_StorageError(t *testing.T) {
	content := []byte("data")

	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			return io.NopCloser(bytes.NewReader(content)), &provider.EntryMeta{
				Name:     "file.txt",
				MimeType: "text/plain",
				Size:     int64(len(content)),
			}, nil, nil
		},
	}

	bucket := &mockBucket{
		onPut: func(_ context.Context, _ string, _ []byte, _ *grub.ObjectInfo) error {
			return errors.New("storage unavailable")
		},
	}

	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from storage write")
	}
}

func TestFetcher_HandleFetch_JobCreateError(t *testing.T) {
	content := []byte("data")

	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
		onCreateJob: func(_ context.Context, _, _, _ string) (*models.Job, error) {
			return nil, errors.New("db error")
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			return io.NopCloser(bytes.NewReader(content)), &provider.EntryMeta{
				Name:     "file.txt",
				MimeType: "text/plain",
				Size:     int64(len(content)),
			}, nil, nil
		},
	}

	bucket := &mockBucket{}
	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from job creation")
	}
}

func TestFetcher_HandleFetch_GetProviderError(t *testing.T) {
	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, errors.New("provider not found")
		},
	}

	prov := &mockFetchProvider{providerType: "test_provider"}
	bucket := &mockBucket{}
	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from GetProvider")
	}
}

func TestFetcher_HandleFetch_RegistryLookupError(t *testing.T) {
	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "unknown_provider"}, nil
		},
	}

	prov := &mockFetchProvider{providerType: "test_provider"}
	bucket := &mockBucket{}
	fetcher := newTestFetcher(store, prov, &provider.Credentials{AccessToken: "test-token"}, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from registry lookup")
	}
}

func TestFetcher_HandleFetch_CredentialLoadError(t *testing.T) {
	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, errors.New("cred store error")
		},
	}

	prov := &mockFetchProvider{providerType: "test_provider"}
	registry := provider.NewRegistry()
	registry.Register(prov)
	credMgr := NewCredentialManager(provStore)
	bucket := &mockBucket{}

	fetcher := NewFetcher(store, credMgr, registry, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from credential loading")
	}
}

func TestFetcher_HandleFetch_CredentialUpdateError(t *testing.T) {
	content := []byte("data")

	refreshedCreds := &provider.Credentials{
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
	}

	provStore := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: `{"access_token":"old"}`}, nil
		},
		onUpdateProviderCredentials: func(_ context.Context, _, _, _ string) error {
			return errors.New("creds persist error")
		},
	}

	store := &mockFetchStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: "test_provider"}, nil
		},
	}

	prov := &mockFetchProvider{
		providerType: "test_provider",
		onFetch: func(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
			return io.NopCloser(bytes.NewReader(content)), &provider.EntryMeta{
				Name:     "file.txt",
				MimeType: "text/plain",
				Size:     int64(len(content)),
			}, refreshedCreds, nil
		},
	}

	registry := provider.NewRegistry()
	registry.Register(prov)
	credMgr := NewCredentialManager(provStore)
	bucket := &mockBucket{}

	fetcher := NewFetcher(store, credMgr, registry, bucket)

	msg := events.FetchMessage{
		VersionID:  "ver-1",
		DocumentID: "doc-1",
		ProviderID: "p-1",
		TenantID:   "t-1",
		Ref:        "ref-1",
		ObjectKey:  "objects/t-1/doc-1",
	}

	err := fetcher.HandleFetch(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from credential update")
	}
}
