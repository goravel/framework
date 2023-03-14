package access

type ResponseImpl struct {
	allowed bool
	message string
}

func (r *ResponseImpl) Allowed() bool {
	return r.allowed
}

func (r *ResponseImpl) Message() string {
	return r.message
}
