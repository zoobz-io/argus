package models

import "fmt"

// DocumentAnalysis holds the structured output from LLM analysis of document content.
type DocumentAnalysis struct {
	Summary  string   `json:"summary"`
	Language string   `json:"language"`
	Topics   []string `json:"topics"`
	Tags     []string `json:"tags"`
}

// Validate satisfies the zyn.Validator interface.
func (a DocumentAnalysis) Validate() error {
	if a.Summary == "" {
		return fmt.Errorf("summary is required")
	}
	if a.Language == "" {
		return fmt.Errorf("language is required")
	}
	return nil
}
