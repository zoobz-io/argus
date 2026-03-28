//go:build testing

package stores

import (
	"testing"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

func TestOffsetPage_PageSize_UsedInResult(t *testing.T) {
	page := models.OffsetPage{Limit: 0}
	if page.PageSize() != models.DefaultPageSize {
		t.Errorf("PageSize: got %d, want %d", page.PageSize(), models.DefaultPageSize)
	}
}

func TestNew(t *testing.T) {
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	mock := soytesting.NewMockDB(t)
	searchMock := &mockSearchProvider{}

	s := New(mock.DB(), astqlpg.New(), nil, searchMock)
	if s.Tenants == nil {
		t.Error("Tenants is nil")
	}
	if s.DocumentVersionSearch == nil {
		t.Error("DocumentVersionSearch is nil")
	}
	if s.Jobs == nil {
		t.Error("Jobs is nil")
	}
}
