package ai

import "sync"

type providerState struct {
	mu   sync.RWMutex
	data map[string]any
}

func newProviderState() *providerState {
	return &providerState{}
}

func (r *providerState) Get(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.data == nil {
		return nil
	}

	return r.data[key]
}

func (r *providerState) Set(key string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data == nil {
		r.data = make(map[string]any)
	}

	if value == nil {
		delete(r.data, key)
		if len(r.data) == 0 {
			r.data = nil
		}
		return
	}

	r.data[key] = value
}
