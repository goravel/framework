package http

type FormRequest interface {
	Messages() map[string]string
	Authorize() bool
}
