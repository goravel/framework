package openai

import contractsai "github.com/goravel/framework/contracts/ai"

type response struct {
	text      string
	usage     *usage
	toolCalls []contractsai.ToolCall
}

type storedFileResponse struct {
	id string
}

func (r *response) Text() string                      { return r.text }
func (r *response) Usage() contractsai.Usage          { return r.usage }
func (r *response) ToolCalls() []contractsai.ToolCall { return r.toolCalls }
func (r *response) Then(callback func(contractsai.Response)) contractsai.Response {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *storedFileResponse) ID() string { return r.id }

type usage struct{ input, output, total int }

func (r *usage) Input() int  { return r.input }
func (r *usage) Output() int { return r.output }
func (r *usage) Total() int  { return r.total }
