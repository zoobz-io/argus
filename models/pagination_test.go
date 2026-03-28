package models

import "testing"

func TestOffsetPage_PageSize(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"zero returns default", 0, DefaultPageSize},
		{"negative returns default", -1, DefaultPageSize},
		{"within range", 50, 50},
		{"at minimum boundary", 1, 1},
		{"at max boundary", MaxPageSize, MaxPageSize},
		{"above max returns max", MaxPageSize + 1, MaxPageSize},
		{"way above max returns max", 9999, MaxPageSize},
		{"default page size value", DefaultPageSize, DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := OffsetPage{Limit: tt.limit}
			if got := p.PageSize(); got != tt.want {
				t.Errorf("PageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
