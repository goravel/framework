package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func TestFailoverError(t *testing.T) {
	cause := assert.AnError
	err := NewFailoverError("openai", contractsai.FailoverReasonRateLimited, cause)

	var failoverErr contractsai.FailoverError
	require.ErrorAs(t, err, &failoverErr)
	assert.Equal(t, contractsai.FailoverReasonRateLimited, failoverErr.Reason())
	assert.Equal(t, "openai", failoverErr.Provider())
	assert.ErrorIs(t, err, cause)
	assert.Equal(t, "ai: provider openai was rate limited", err.Error())
}

func TestNewFailoverProvider(t *testing.T) {
	primaryProvider := &failoverTestProvider{}
	backupProvider := &failoverTestProvider{}

	assert.Same(t, primaryProvider, newFailoverProvider([]resolvedProvider{{name: "primary", provider: primaryProvider}}))

	providers := []resolvedProvider{
		{name: "primary", provider: primaryProvider},
		{name: "backup", provider: backupProvider},
	}
	provider, ok := newFailoverProvider(providers).(*failoverProvider)
	require.True(t, ok)

	providers[0].provider = backupProvider
	assert.Same(t, primaryProvider, provider.providers[0].provider)
	assert.Same(t, backupProvider, provider.providers[1].provider)
}

func TestFailoverProviderStreamSuppressesPendingFailoverError(t *testing.T) {
	failoverErr := NewFailoverError("primary", contractsai.FailoverReasonRateLimited, assert.AnError)
	primaryProvider := &failoverTestProvider{
		streamEvents: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeError, Error: "rate limited"}},
		streamErr:    failoverErr,
	}
	backupProvider := &failoverTestProvider{
		streamEvents: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"}},
		streamResp:   &failoverTestResponse{text: "backup"},
	}
	provider := &failoverProvider{providers: []resolvedProvider{
		{name: "primary", provider: primaryProvider},
		{name: "backup", provider: backupProvider},
	}}

	stream, err := provider.Stream(context.Background(), contractsai.AgentPrompt{})
	require.NoError(t, err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"}}, events)
	assert.Equal(t, 1, primaryProvider.streamCalls)
	assert.Equal(t, 1, backupProvider.streamCalls)
}

func TestFailoverProviderStreamEmitsPendingErrorBeforeNonFailoverError(t *testing.T) {
	primaryProvider := &failoverTestProvider{
		streamEvents: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeError, Error: "invalid request"}},
		streamErr:    assert.AnError,
	}
	backupProvider := &failoverTestProvider{streamResp: &failoverTestResponse{text: "backup"}}
	provider := &failoverProvider{providers: []resolvedProvider{
		{name: "primary", provider: primaryProvider},
		{name: "backup", provider: backupProvider},
	}}

	stream, err := provider.Stream(context.Background(), contractsai.AgentPrompt{})
	require.NoError(t, err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeError, Error: "invalid request"}}, events)
	assert.Equal(t, 1, primaryProvider.streamCalls)
	assert.Zero(t, backupProvider.streamCalls)
}

func TestScopedProviderState(t *testing.T) {
	state := newProviderState()
	scoped := scopedProviderState{provider: "openai", state: state}

	scoped.Set("response_id", "resp_123")

	assert.Nil(t, state.Get("response_id"))
	assert.Equal(t, "resp_123", state.Get("openai:response_id"))
	assert.Equal(t, "resp_123", scoped.Get("response_id"))

	scoped.Set("response_id", nil)
	assert.Nil(t, scoped.Get("response_id"))
	assert.Nil(t, state.Get("openai:response_id"))
}

type failoverTestProvider struct {
	promptResp   contractsai.AgentResponse
	promptErr    error
	streamResp   contractsai.AgentResponse
	streamErr    error
	streamEvents []contractsai.StreamEvent
	streamCalls  int
}

func (p *failoverTestProvider) Prompt(context.Context, contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	return p.promptResp, p.promptErr
}

func (p *failoverTestProvider) Stream(ctx context.Context, _ contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	p.streamCalls++

	return NewStreamableResponse(ctx, func(_ context.Context, emit func(contractsai.StreamEvent) error) (contractsai.AgentResponse, error) {
		for _, event := range p.streamEvents {
			if err := emit(event); err != nil {
				return nil, err
			}
		}

		return p.streamResp, p.streamErr
	}), nil
}

type failoverTestResponse struct {
	text string
}

func (r *failoverTestResponse) Text() string { return r.text }

func (r *failoverTestResponse) Usage() contractsai.Usage { return nil }

func (r *failoverTestResponse) ToolCalls() []contractsai.ToolCall { return nil }

func (r *failoverTestResponse) Then(callback func(contractsai.AgentResponse)) contractsai.AgentResponse {
	if callback != nil {
		callback(r)
	}

	return r
}
