package http

type Json map[string]interface{}

//go:generate mockery --name=Response
type Response interface {
	String(code int, format string, values ...interface{})
	Json(code int, obj interface{})
	File(filepath string)
	Download(filepath, filename string)
	Success() ResponseSuccess
	Header(key, value string) Response
}

//go:generate mockery --name=ResponseSuccess
type ResponseSuccess interface {
	String(format string, values ...interface{})
	Json(obj interface{})
}
