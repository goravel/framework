package ai

import contractsai "github.com/goravel/framework/contracts/ai"

type middlewareResponse struct {
	response  contractsai.Response
	callbacks []func(contractsai.Response) error
	err       error
	resolved  bool
	usedThen  bool
}

func newResolvedMiddlewareResponse(response contractsai.Response) *middlewareResponse {
	return &middlewareResponse{
		response: response,
		resolved: true,
	}
}

func newDeferredMiddlewareResponse() *middlewareResponse {
	return &middlewareResponse{}
}

func (r *middlewareResponse) Text() string {
	if r.response == nil {
		return ""
	}

	return r.response.Text()
}

func (r *middlewareResponse) Usage() contractsai.Usage {
	if r.response == nil {
		return nil
	}

	return r.response.Usage()
}

func (r *middlewareResponse) ToolCalls() []contractsai.ToolCall {
	if r.response == nil {
		return nil
	}

	return r.response.ToolCalls()
}

func (r *middlewareResponse) Then(callback func(contractsai.Response) error) contractsai.Response {
	if callback == nil || r.err != nil {
		return r
	}

	r.usedThen = true

	if !r.resolved {
		r.callbacks = append(r.callbacks, callback)
		return r
	}

	r.err = callback(r)

	return r
}

func (r *middlewareResponse) Resolve(response contractsai.Response) error {
	r.response = response
	r.resolved = true

	for _, callback := range r.callbacks {
		if err := callback(r); err != nil {
			r.err = err
			return err
		}
	}

	r.callbacks = nil

	return nil
}

func (r *middlewareResponse) Err() error {
	return r.err
}

func (r *middlewareResponse) Unwrap() contractsai.Response {
	if r == nil || r.response == nil {
		return nil
	}

	if r.usedThen || r.err != nil {
		return r
	}

	return r.response
}
