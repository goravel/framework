package openai

import contractsai "github.com/goravel/framework/contracts/ai"

type response struct {
	text  string
	usage *usage
}

func (r *response) Text() string             { return r.text }
func (r *response) Usage() contractsai.Usage { return r.usage }

type usage struct{ input, output, total int }

func (r *usage) Input() int  { return r.input }
func (r *usage) Output() int { return r.output }
func (r *usage) Total() int  { return r.total }
