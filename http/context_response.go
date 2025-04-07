package http

import (
	"bytes"
	"net/http"
	"sync"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
)

var contextResponsePool = sync.Pool{New: func() any {
	return &ContextResponse{}
}}

type ContextResponse struct {
	w      http.ResponseWriter
	r      *http.Request
	origin contractshttp.ResponseOrigin
}

func NewContextResponse(w http.ResponseWriter, r *http.Request, origin contractshttp.ResponseOrigin) contractshttp.ContextResponse {
	response := contextResponsePool.Get().(*ContextResponse)
	response.w = w
	response.r = r
	response.origin = origin
	return response
}

func (r *ContextResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	if cookie.MaxAge == 0 {
		if !cookie.Expires.IsZero() {
			cookie.MaxAge = int(cookie.Expires.Sub(carbon.Now().StdTime()).Seconds())
		}
	}

	sameSiteOptions := map[string]http.SameSite{
		"strict": http.SameSiteStrictMode,
		"lax":    http.SameSiteLaxMode,
		"none":   http.SameSiteNoneMode,
	}
	sameSite, ok := sameSiteOptions[cookie.SameSite]
	if !ok {
		sameSite = http.SameSiteDefaultMode
	}

	http.SetCookie(r.w, &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Expires:  cookie.Expires,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
		MaxAge:   cookie.MaxAge,
		SameSite: sameSite,
	})

	return r
}

func (r *ContextResponse) Data(code int, contentType string, data []byte) contractshttp.AbortableResponse {
	return &DataResponse{code, contentType, data, r.w}
}

func (r *ContextResponse) Download(filepath, filename string) contractshttp.Response {
	return &DownloadResponse{filename, filepath, r.w, r.r}
}

func (r *ContextResponse) File(filepath string) contractshttp.Response {
	return &FileResponse{filepath, r.w, r.r}
}

func (r *ContextResponse) Header(key, value string) contractshttp.ContextResponse {
	r.w.Header().Set(key, value)

	return r
}

func (r *ContextResponse) Json(code int, obj any) contractshttp.AbortableResponse {
	return &JsonResponse{code, obj, r.w}
}

func (r *ContextResponse) NoContent(code ...int) contractshttp.AbortableResponse {
	if len(code) > 0 {
		return &NoContentResponse{code[0], r.w}
	}

	return &NoContentResponse{http.StatusNoContent, r.w}
}

func (r *ContextResponse) Origin() contractshttp.ResponseOrigin {
	return r.origin
}

func (r *ContextResponse) Redirect(code int, location string) contractshttp.AbortableResponse {
	return &RedirectResponse{code, location, r.w, r.r}
}

func (r *ContextResponse) String(code int, format string, values ...any) contractshttp.AbortableResponse {
	return &StringResponse{code, format, r.w, values}
}

func (r *ContextResponse) Success() contractshttp.ResponseStatus {
	return NewStatus(r.w, http.StatusOK)
}

func (r *ContextResponse) Status(code int) contractshttp.ResponseStatus {
	return NewStatus(r.w, code)
}

func (r *ContextResponse) Stream(code int, step func(w contractshttp.StreamWriter) error) contractshttp.Response {
	return &StreamResponse{code, r.w, step}
}

func (r *ContextResponse) View() contractshttp.ResponseView {
	panic("not implemented")
}

func (r *ContextResponse) WithoutCookie(name string) contractshttp.ContextResponse {
	http.SetCookie(r.w, &http.Cookie{
		Name:   name,
		MaxAge: -1,
	})

	return r
}

func (r *ContextResponse) Writer() http.ResponseWriter {
	return r.w
}

func (r *ContextResponse) Flush() {
	if flusher, ok := r.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

type Status struct {
	w      http.ResponseWriter
	status int
}

func NewStatus(w http.ResponseWriter, code int) *Status {
	return &Status{w, code}
}

func (r *Status) Data(contentType string, data []byte) contractshttp.AbortableResponse {
	return &DataResponse{r.status, contentType, data, r.w}
}

func (r *Status) Json(obj any) contractshttp.AbortableResponse {
	return &JsonResponse{r.status, obj, r.w}
}

func (r *Status) String(format string, values ...any) contractshttp.AbortableResponse {
	return &StringResponse{r.status, format, r.w, values}
}

func (r *Status) Stream(step func(w contractshttp.StreamWriter) error) contractshttp.Response {
	return &StreamResponse{r.status, r.w, step}
}

type ResponseOrigin struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *ResponseOrigin) Size() int {
	return w.body.Len()
}

// Status TODO not support in http.ResponseWriter
func (w *ResponseOrigin) Status() int {
	return 0
}

func (w *ResponseOrigin) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func (w *ResponseOrigin) WriteString(s string) (int, error) {
	w.body.WriteString(s)

	return w.ResponseWriter.Write([]byte(s))
}

func (w *ResponseOrigin) Body() *bytes.Buffer {
	return w.body
}

func (w *ResponseOrigin) Header() http.Header {
	return w.ResponseWriter.Header()
}
