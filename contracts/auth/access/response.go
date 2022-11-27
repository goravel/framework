package access

type Response struct {
	allowed bool
	message string
}

func NewAllowResponse() *Response {
	return &Response{allowed: true}
}

func NewDenyResponse(message string) *Response {
	return &Response{message: message}
}

func (r Response) Allowed() bool {
	return r.allowed
}

func (r Response) Message() string {
	return r.message
}
