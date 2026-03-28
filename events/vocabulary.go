package events

import "github.com/zoobz-io/capitan"

// Vocabulary pipeline signals.
var (
	VocabularyClassifyUnavailable = capitan.NewSignal("argus.vocabulary.classify.unavailable", "Injection classifier unavailable for vocabulary check")
	VocabularyClassifyRejected    = capitan.NewSignal("argus.vocabulary.classify.rejected", "Vocabulary entry flagged as prompt injection")
)

// Vocabulary field keys.
var (
	VocabularyNameKey = capitan.NewStringKey("vocabulary_name")
)
