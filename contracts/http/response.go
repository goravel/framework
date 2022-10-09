package http

type Json map[string]interface{}

type Response interface {
	String(code int, format string, values ...interface{})
	Json(code int, obj interface{})
	File(filepath string)
	Download(filepath, filename string)
	Success() ResponseSuccess
	Header(key, value string) Response
}

type ResponseSuccess interface {
	String(format string, values ...interface{})
	Json(obj interface{})
}
