package handlers

import (
	"testing"

	"github.com/zoobz-io/rocco"
)

func TestOffsetPageFromQuery_Defaults(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{}}
	page := offsetPageFromQuery(params)
	if page.Limit != 20 {
		t.Errorf("expected default 20, got %d", page.Limit)
	}
	if page.Offset != 0 {
		t.Error("expected zero offset")
	}
}

func TestOffsetPageFromQuery_WithValues(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{"limit": "50", "offset": "10"}}
	page := offsetPageFromQuery(params)
	if page.Limit != 50 {
		t.Errorf("expected 50, got %d", page.Limit)
	}
	if page.Offset != 10 {
		t.Errorf("expected offset 10, got %d", page.Offset)
	}
}

func TestOffsetPageFromQuery_InvalidLimit(t *testing.T) {
	params := &rocco.Params{Query: map[string]string{"limit": "bad"}}
	page := offsetPageFromQuery(params)
	if page.Limit != 20 {
		t.Errorf("expected default on invalid, got %d", page.Limit)
	}
}

func TestPathID(t *testing.T) {
	params := &rocco.Params{Path: map[string]string{"id": "abc"}}
	if got := pathID(params, "id"); got != "abc" {
		t.Errorf("expected abc, got %q", got)
	}
}
