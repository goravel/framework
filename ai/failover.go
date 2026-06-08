package ai

import (
	"context"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

type resolvedProvider struct {
	name     string
	provider contractsai.Provider
}

type failoverError struct {
	provider string
	reason   contractsai.FailoverReason
	cause    error
}

type failoverProvider struct {
	providers []resolvedProvider
}

type scopedProviderState struct {
	provider string
	state    contractsai.ProviderState
}

var _ contractsai.FailoverError = (*failoverError)(nil)

func NewFailoverError(provider string, reason contractsai.FailoverReason, cause error) error {
	return &failoverError{provider: provider, reason: reason, cause: cause}
}

func newFailoverProvider(providers []resolvedProvider) contractsai.Provider {
	if len(providers) == 1 {
		return providers[0].provider
	}

	return &failoverProvider{providers: append([]resolvedProvider(nil), providers...)}
}

func (e *failoverError) Error() string {
	switch e.reason {
	case contractsai.FailoverReasonRateLimited:
		return errors.AIFailoverRateLimited.Args(e.provider).Error()
	case contractsai.FailoverReasonProviderOverloaded:
		return errors.AIFailoverProviderOverloaded.Args(e.provider).Error()
	case contractsai.FailoverReasonInsufficientCredits:
		return errors.AIFailoverInsufficientCredits.Args(e.provider).Error()
	default:
		if e.cause != nil {
			return e.cause.Error()
		}

		return errors.AIProviderNotSupported.Args(e.provider).Error()
	}
}

func (e *failoverError) Reason() contractsai.FailoverReason {
	return e.reason
}

func (e *failoverError) Provider() string {
	return e.provider
}

func (e *failoverError) Unwrap() error {
	return e.cause
}

func (r *failoverProvider) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	var lastErr error
	for _, resolvedProvider := range r.providers {
		response, err := resolvedProvider.provider.Prompt(ctx, r.promptFor(resolvedProvider, prompt))
		if err == nil {
			return response, nil
		}
		if !isFailoverError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, lastErr
}

func (r *failoverProvider) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return NewStreamableResponse(ctx, func(streamCtx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.AgentResponse, error) {
		var lastErr error
		for _, resolvedProvider := range r.providers {
			stream, err := resolvedProvider.provider.Stream(streamCtx, r.promptFor(resolvedProvider, prompt))
			if err != nil {
				if !isFailoverError(err) {
					return nil, err
				}

				lastErr = err
				continue
			}
			if stream == nil {
				return nil, errors.AIResponseIsNil
			}

			response, started, err := r.forwardStream(stream, emit)
			if err == nil {
				return response, nil
			}
			if !isFailoverError(err) || started {
				return nil, err
			}

			lastErr = err
		}

		return nil, lastErr
	}), nil
}

func (r *failoverProvider) promptFor(provider resolvedProvider, prompt contractsai.AgentPrompt) contractsai.AgentPrompt {
	if prompt.ProviderState != nil {
		prompt.ProviderState = scopedProviderState{provider: provider.name, state: prompt.ProviderState}
	}

	return prompt
}

func (r *failoverProvider) forwardStream(stream contractsai.StreamableAgentResponse, emit func(contractsai.StreamEvent) error) (contractsai.AgentResponse, bool, error) {
	var response contractsai.AgentResponse
	stream.Then(func(resp contractsai.AgentResponse) {
		response = resp
	})

	started := false
	var pendingErrors []contractsai.StreamEvent
	err := stream.Each(func(event contractsai.StreamEvent) error {
		if event.Type == contractsai.StreamEventTypeError && !started {
			pendingErrors = append(pendingErrors, event)
			return nil
		}

		started = true
		if err := emitPendingStreamErrors(pendingErrors, emit); err != nil {
			return err
		}
		pendingErrors = nil

		return emit(event)
	})
	if err != nil {
		if isFailoverError(err) && !started {
			return nil, false, err
		}
		if emitErr := emitPendingStreamErrors(pendingErrors, emit); emitErr != nil {
			return nil, started, emitErr
		}

		return response, started, err
	}
	if err := emitPendingStreamErrors(pendingErrors, emit); err != nil {
		return nil, started, err
	}

	return response, started, nil
}

func (s scopedProviderState) Get(key string) any {
	return s.state.Get(s.key(key))
}

func (s scopedProviderState) Set(key string, value any) {
	s.state.Set(s.key(key), value)
}

func (s scopedProviderState) key(key string) string {
	if s.provider == "" {
		return key
	}

	return s.provider + ":" + key
}

func emitPendingStreamErrors(events []contractsai.StreamEvent, emit func(contractsai.StreamEvent) error) error {
	for _, event := range events {
		if err := emit(event); err != nil {
			return err
		}
	}

	return nil
}

func isFailoverError(err error) bool {
	var failoverErr contractsai.FailoverError
	return errors.As(err, &failoverErr)
}
