package models

import "testing"

func TestDocumentAnalysis_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   DocumentAnalysis
		wantErr string
	}{
		{
			name:    "valid",
			input:   DocumentAnalysis{Summary: "a summary", Language: "en"},
			wantErr: "",
		},
		{
			name:    "empty summary",
			input:   DocumentAnalysis{Summary: "", Language: "en"},
			wantErr: "summary is required",
		},
		{
			name:    "empty language",
			input:   DocumentAnalysis{Summary: "a summary", Language: ""},
			wantErr: "language is required",
		},
		{
			name:    "both empty returns summary error first",
			input:   DocumentAnalysis{},
			wantErr: "summary is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
