package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

type FailoverTestSuite struct {
	suite.Suite
}

func TestFailoverTestSuite(t *testing.T) {
	suite.Run(t, &FailoverTestSuite{})
}

func (s *FailoverTestSuite) TestFailoverError() {
	cause := assert.AnError
	err := NewFailoverError("openai", contractsai.FailoverReasonRateLimited, cause)

	var failoverErr contractsai.FailoverError
	s.Require().ErrorAs(err, &failoverErr)
	s.Equal(contractsai.FailoverReasonRateLimited, failoverErr.Reason())
	s.Equal("openai", failoverErr.Provider())
	s.ErrorIs(err, cause)
	s.Equal("ai: provider openai was rate limited", err.Error())
}

func (s *FailoverTestSuite) TestFailoverErrorCustomReason() {
	cause := aiTestError("maximum context length exceeded")
	err := NewFailoverError("openai", contractsai.FailoverReason("context_length_exceeded"), cause)

	var failoverErr contractsai.FailoverError
	s.Require().ErrorAs(err, &failoverErr)
	s.Equal(contractsai.FailoverReason("context_length_exceeded"), failoverErr.Reason())
	s.Equal("openai", failoverErr.Provider())
	s.ErrorIs(err, cause)
	s.Equal("ai: provider openai failed over because context_length_exceeded", err.Error())
}

func (s *FailoverTestSuite) TestNewFailoverProvider() {
	primaryProvider := &failoverTestProvider{}
	backupProvider := &failoverTestProvider{}

	s.Same(primaryProvider, newFailoverProvider([]resolvedProvider{{name: "primary", provider: primaryProvider}}))

	providers := []resolvedProvider{
		{name: "primary", provider: primaryProvider},
		{name: "backup", provider: backupProvider},
	}
	provider, ok := newFailoverProvider(providers).(*failoverProvider)
	s.Require().True(ok)

	providers[0].provider = backupProvider
	s.Same(primaryProvider, provider.providers[0].provider)
	s.Same(backupProvider, provider.providers[1].provider)
}

func (s *FailoverTestSuite) TestStreamSuppressesPendingFailoverError() {
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
	s.Require().NoError(err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	s.Require().NoError(err)
	s.Equal([]contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"}}, events)
	s.Equal(1, primaryProvider.streamCalls)
	s.Equal(1, backupProvider.streamCalls)
}

func (s *FailoverTestSuite) TestStreamSuppressesPendingConfiguredFailoverError() {
	failoverRules, err := newFailoverRules("primary", map[contractsai.FailoverReason][]string{
		"context_length_exceeded": {"context length"},
	})
	s.Require().NoError(err)

	primaryProvider := &failoverTestProvider{
		streamEvents: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeError, Error: "context length exceeded"}},
		streamErr:    aiTestError("maximum context length exceeded"),
	}
	backupProvider := &failoverTestProvider{
		streamEvents: []contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"}},
		streamResp:   &failoverTestResponse{text: "backup"},
	}
	provider := &failoverProvider{providers: []resolvedProvider{
		{name: "primary", provider: primaryProvider, failoverRules: failoverRules},
		{name: "backup", provider: backupProvider},
	}}

	stream, err := provider.Stream(context.Background(), contractsai.AgentPrompt{})
	s.Require().NoError(err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	s.Require().NoError(err)
	s.Equal([]contractsai.StreamEvent{{Type: contractsai.StreamEventTypeTextDelta, Delta: "backup"}}, events)
	s.Equal(1, primaryProvider.streamCalls)
	s.Equal(1, backupProvider.streamCalls)
}

func (s *FailoverTestSuite) TestStreamEmitsPendingErrorBeforeNonFailoverError() {
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
	s.Require().NoError(err)

	var events []contractsai.StreamEvent
	err = stream.Each(func(event contractsai.StreamEvent) error {
		events = append(events, event)
		return nil
	})

	s.Equal(assert.AnError, err)
	s.Equal([]contractsai.StreamEvent{{Type: contractsai.StreamEventTypeError, Error: "invalid request"}}, events)
	s.Equal(1, primaryProvider.streamCalls)
	s.Zero(backupProvider.streamCalls)
}

func (s *FailoverTestSuite) TestScopedProviderState() {
	state := newProviderState()
	scoped := scopedProviderState{provider: "openai", state: state}

	scoped.Set("response_id", "resp_123")

	s.Nil(state.Get("response_id"))
	s.Equal("resp_123", state.Get("openai:response_id"))
	s.Equal("resp_123", scoped.Get("response_id"))

	scoped.Set("response_id", nil)
	s.Nil(scoped.Get("response_id"))
	s.Nil(state.Get("openai:response_id"))
}

func (s *FailoverTestSuite) TestFailoverRulesMatchSubstringAndRegex() {
	failoverRules, err := newFailoverRules("openai", map[contractsai.FailoverReason][]string{
		"context_length_exceeded": {"context length"},
		"model_overloaded":        {"/(?i)model.*overloaded/"},
	})
	s.Require().NoError(err)
	provider := resolvedProvider{name: "openai", failoverRules: failoverRules}

	contextErr := aiTestError("maximum context length exceeded")
	err = provider.failoverError(contextErr)
	var failoverErr contractsai.FailoverError
	s.Require().ErrorAs(err, &failoverErr)
	s.Equal(contractsai.FailoverReason("context_length_exceeded"), failoverErr.Reason())
	s.ErrorIs(err, contextErr)

	overloadedErr := aiTestError("the MODEL is overloaded")
	err = provider.failoverError(overloadedErr)
	s.Require().ErrorAs(err, &failoverErr)
	s.Equal(contractsai.FailoverReason("model_overloaded"), failoverErr.Reason())
	s.ErrorIs(err, overloadedErr)
}

func (s *FailoverTestSuite) TestFailoverRulesReturnErrorForInvalidRegex() {
	_, err := newFailoverRules("openai", map[contractsai.FailoverReason][]string{
		"bad_pattern": {"/[/"},
	})

	s.ErrorIs(err, errors.AIFailoverPatternInvalid)
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

type aiTestError string

func (e aiTestError) Error() string {
	return string(e)
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
