package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func TestNewTranscriptionResponse(t *testing.T) {
	segments := []contractsai.TranscriptionSegment{{
		Speaker: "speaker-1",
		Start:   time.Second,
		End:     2 * time.Second,
		Text:    "hello",
	}}
	response := NewTranscriptionResponse("hello", segments, NewUsage(1, 2, 3))

	assert.Equal(t, "hello", response.Text())
	assert.Equal(t, segments, response.Segments())
	assert.Equal(t, 1, response.Usage().Input())
	assert.Equal(t, 2, response.Usage().Output())
	assert.Equal(t, 3, response.Usage().Total())
}

func TestTranscriptionResponse_SegmentsCloned(t *testing.T) {
	response := &transcriptionResponse{segments: []contractsai.TranscriptionSegment{{Text: "hello"}}}

	segments := response.Segments()
	segments[0].Text = "changed"

	assert.Equal(t, "hello", response.segments[0].Text)
}

func TestTranscriptionResponse_Then(t *testing.T) {
	response := &transcriptionResponse{text: "hello"}
	called := false

	result := response.Then(func(result contractsai.TranscriptionResponse) {
		called = true
		assert.Equal(t, "hello", result.Text())
	})

	assert.True(t, called)
	assert.Same(t, response, result)
}
