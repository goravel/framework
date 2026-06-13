package ai

import (
	"context"
	"regexp"
	"sort"
	"strings"

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

// FailoverRules matches provider errors against configured failover patterns.
type FailoverRules struct {
	rules []failoverRule
}

type failoverRule struct {
	reason  contractsai.FailoverReason
	pattern string
	regex   *regexp.Regexp
}

type failoverProvider struct {
	providers []resolvedProvider
}

type scopedProviderState struct {
	provider string
	state    contractsai.ProviderState
}

var _ contractsai.FailoverError = (*failoverError)(nil)

// NewFailoverError returns an error that instructs the framework to try the next provider.
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
	if e.reason != "" {
		return errors.AIFailoverReason.Args(e.provider, e.reason).Error()
	}

	if e.cause != nil {
		return e.cause.Error()
	}

	return errors.AIProviderNotSupported.Args(e.provider).Error()
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

// NewFailoverRules compiles provider failover patterns.
// Plain patterns use substring matching; slash-delimited patterns use Go regex syntax.
func NewFailoverRules(provider string, patterns map[contractsai.FailoverReason][]string) (FailoverRules, error) {
	if len(patterns) == 0 {
		return FailoverRules{}, nil
	}

	reasons := make([]string, 0, len(patterns))
	for reason := range patterns {
		if reason != "" {
			reasons = append(reasons, string(reason))
		}
	}
	sort.Strings(reasons)

	var rules []failoverRule
	for _, reasonValue := range reasons {
		reason := contractsai.FailoverReason(reasonValue)
		for _, pattern := range patterns[reason] {
			if pattern == "" {
				continue
			}

			rule := failoverRule{reason: reason, pattern: pattern}
			if regexPattern, ok := failoverRegexPattern(pattern); ok {
				if regexPattern == "" {
					continue
				}

				regex, err := regexp.Compile(regexPattern)
				if err != nil {
					return FailoverRules{}, errors.AIFailoverPatternInvalid.Args(provider, reason, pattern, err)
				}

				rule.regex = regex
			}

			rules = append(rules, rule)
		}
	}

	return FailoverRules{rules: rules}, nil
}

func failoverRegexPattern(pattern string) (string, bool) {
	if len(pattern) < 2 || !strings.HasPrefix(pattern, "/") || !strings.HasSuffix(pattern, "/") {
		return "", false
	}

	return pattern[1 : len(pattern)-1], true
}

// Match returns the configured failover reason for err.
func (r FailoverRules) Match(err error) (contractsai.FailoverReason, bool) {
	if err == nil {
		return "", false
	}

	message := err.Error()
	for _, rule := range r.rules {
		if rule.matches(message) {
			return rule.reason, true
		}
	}

	return "", false
}

// Wrap converts err to a failover error when it matches a configured rule.
func (r FailoverRules) Wrap(provider string, err error) error {
	if err == nil || isFailoverError(err) {
		return err
	}

	if reason, ok := r.Match(err); ok {
		return NewFailoverError(provider, reason, err)
	}

	return err
}

func (r failoverRule) matches(message string) bool {
	if r.regex != nil {
		return r.regex.MatchString(message)
	}

	return strings.Contains(message, r.pattern)
}

func isFailoverError(err error) bool {
	var failoverErr contractsai.FailoverError
	return errors.As(err, &failoverErr)
}
