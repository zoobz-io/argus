package wire

import "testing"

func TestOffsetParams_Validate_Valid(t *testing.T) {
	for _, limit := range []int{1, 20, 100} {
		p := OffsetParams{Limit: limit}
		if err := p.Validate(); err != nil {
			t.Errorf("expected valid for limit %d, got: %v", limit, err)
		}
	}
}

func TestOffsetParams_Validate_Invalid(t *testing.T) {
	for _, limit := range []int{0, -1, 101} {
		p := OffsetParams{Limit: limit}
		if err := p.Validate(); err == nil {
			t.Errorf("expected error for limit %d", limit)
		}
	}
}

func TestOffsetParams_Validate_NegativeOffset(t *testing.T) {
	p := OffsetParams{Offset: -1, Limit: 20}
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative offset")
	}
}

func TestOffsetParams_Clone(t *testing.T) {
	orig := OffsetParams{Offset: 10, Limit: 20}
	cloned := orig.Clone()

	if cloned.Offset != 10 {
		t.Errorf("offset mismatch: %d", cloned.Offset)
	}
	if cloned.Limit != 20 {
		t.Errorf("limit mismatch: %d", cloned.Limit)
	}
}
