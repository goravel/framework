package ai

import "time"

// TranscriptionRequest defines a fluent speech-to-text request.
type TranscriptionRequest interface {
	Model(model string) TranscriptionRequest
	Provider(provider string, failovers ...string) TranscriptionRequest
	Language(language string) TranscriptionRequest
	Diarize() TranscriptionRequest
	Timeout(timeout time.Duration) TranscriptionRequest
	Generate() (TranscriptionResponse, error)
}

// TranscriptionSegment represents a single transcript segment.
// Speaker may be empty when the provider does not return diarization data.
type TranscriptionSegment struct {
	Speaker string
	Start   time.Duration
	End     time.Duration
	Text    string
}
