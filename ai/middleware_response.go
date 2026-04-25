package ai

import contractsai "github.com/goravel/framework/contracts/ai"

type middlewareResponse struct {
	response  contractsai.Response
	callbacks []func(contractsai.Response)
	resolved  bool
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

func (r *middlewareResponse) Then(callback func(contractsai.Response)) contractsai.Response {
	if callback == nil {
		return r
	}

	if !r.resolved {
		r.callbacks = append(r.callbacks, callback)
		return r
	}

	callback(r)

	return r
}

func (r *middlewareResponse) Resolve(response contractsai.Response) {
	r.response = response
	r.resolved = true

	for _, callback := range r.callbacks {
		callback(r)
	}

	r.callbacks = nil
}

func (r *middlewareResponse) Unwrap() contractsai.Response {
	if r == nil || r.response == nil {
		return nil
	}

	return r.response
}
