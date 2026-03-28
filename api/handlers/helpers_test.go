package handlers

import (
	"testing"

	"github.com/zoobz-io/rocco"
)

func TestOffsetPageFromQuery_Defaults(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{}}
	page := offsetPageFromQuery(params)

	if page.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", page.Limit)
	}
	if page.Offset != 0 {
		t.Error("expected zero offset")
	}
}

func TestOffsetPageFromQuery_WithValues(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{"limit": "50", "offset": "10"}}
	page := offsetPageFromQuery(params)

	if page.Limit != 50 {
		t.Errorf("expected limit 50, got %d", page.Limit)
	}
	if page.Offset != 10 {
		t.Errorf("expected offset 10, got %d", page.Offset)
	}
}

func TestOffsetPageFromQuery_InvalidLimit(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{"limit": "notanumber"}}
	page := offsetPageFromQuery(params)

	if page.Limit != 20 {
		t.Errorf("expected default limit on invalid input, got %d", page.Limit)
	}
}

func TestPathID(t *testing.T) {
	params := &rocco.Params{Path: map[string]string{"id": "abc-123"}}
	if got := pathID(params, "id"); got != "abc-123" {
		t.Errorf("expected abc-123, got %q", got)
	}
}

func TestPathID_Missing(t *testing.T) {
	params := &rocco.Params{Path: map[string]string{}}
	if got := pathID(params, "id"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
