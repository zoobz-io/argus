//go:build testing

package wire

import "testing"

func TestOffsetParams_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		limit  int
	}{
		{"min limit", 0, 1},
		{"typical", 0, 20},
		{"max limit", 0, 100},
		{"with offset", 50, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := OffsetParams{Offset: tt.offset, Limit: tt.limit}
			if err := p.Validate(); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestOffsetParams_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		limit  int
	}{
		{"zero limit", 0, 0},
		{"negative limit", 0, -1},
		{"exceeds max limit", 0, 101},
		{"negative offset", -1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := OffsetParams{Offset: tt.offset, Limit: tt.limit}
			if err := p.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestOffsetParams_Clone(t *testing.T) {
	orig := OffsetParams{Offset: 10, Limit: 20}
	cloned := orig.Clone()

	if cloned.Limit != orig.Limit {
		t.Errorf("limit mismatch: %d vs %d", cloned.Limit, orig.Limit)
	}
	if cloned.Offset != orig.Offset {
		t.Errorf("offset mismatch: %d vs %d", cloned.Offset, orig.Offset)
	}
}
