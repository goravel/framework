package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
)

// DataResponse represents a raw data response
type DataResponse struct {
	code        int
	contentType string
	data        []byte
	w           http.ResponseWriter
}

func (r *DataResponse) Render() error {
	r.w.Header().Set("Content-Type", r.contentType)
	r.w.WriteHeader(r.code)
	_, err := r.w.Write(r.data)
	return err
}

func (r *DataResponse) Abort() error {
	// In net/http, aborting just means stopping the handler chain
	// and writing the response immediately
	return r.Render()
}

// DownloadResponse represents a file download response
type DownloadResponse struct {
	filename string
	filepath string
	w        http.ResponseWriter
	r        *http.Request
}

func (r *DownloadResponse) Render() error {
	// Set headers for attachment download
	r.w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", r.filename))
	r.w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(r.w, r.r, r.filepath)
	return nil
}

// FileResponse represents a file display response
type FileResponse struct {
	filepath string
	w        http.ResponseWriter
	r        *http.Request
}

func (r *FileResponse) Render() error {
	http.ServeFile(r.w, r.r, r.filepath)
	return nil
}

// JsonResponse represents a JSON response
type JsonResponse struct {
	code int
	obj  any
	w    http.ResponseWriter
}

func (r *JsonResponse) Render() error {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.code)

	encoder := json.NewEncoder(r.w)
	return encoder.Encode(r.obj)
}

func (r *JsonResponse) Abort() error {
	return r.Render()
}

// NoContentResponse represents an empty response with status code
type NoContentResponse struct {
	code int
	w    http.ResponseWriter
}

func (r *NoContentResponse) Render() error {
	r.w.WriteHeader(r.code)
	return nil
}

func (r *NoContentResponse) Abort() error {
	return r.Render()
}

// RedirectResponse represents a redirect response
type RedirectResponse struct {
	code     int
	location string
	w        http.ResponseWriter
	r        *http.Request
}

func (r *RedirectResponse) Render() error {
	http.Redirect(r.w, r.r, r.location, r.code)
	return nil
}

func (r *RedirectResponse) Abort() error {
	return r.Render()
}

// StringResponse represents a plain text response
type StringResponse struct {
	code   int
	format string
	w      http.ResponseWriter
	values []any
}

func (r *StringResponse) Render() error {
	r.w.Header().Set("Content-Type", "text/plain")
	r.w.WriteHeader(r.code)
	_, err := fmt.Fprintf(r.w, r.format, r.values...)
	return err
}

func (r *StringResponse) Abort() error {
	return r.Render()
}

// HtmlResponse represents an HTML response
type HtmlResponse struct {
	data any
	w    http.ResponseWriter
	view string
}

func (r *HtmlResponse) Render() error {
	r.w.Header().Set("Content-Type", "text/html")
	r.w.WriteHeader(http.StatusOK)

	// Here, we'd need an HTML template rendering mechanism
	// This is a simplified version assuming the view is already HTML content
	if htmlStr, ok := r.data.(string); ok {
		_, err := r.w.Write([]byte(htmlStr))
		return err
	} else {
		jsonData, err := json.Marshal(r.data)
		if err != nil {
			return err
		}
		_, err = r.w.Write(jsonData)
		return err
	}
}

// StreamResponse represents a streaming response
type StreamResponse struct {
	code   int
	w      http.ResponseWriter
	writer func(w contractshttp.StreamWriter) error
}

func (r *StreamResponse) Render() error {
	r.w.Header().Set("Content-Type", "text/event-stream")
	r.w.Header().Set("Cache-Control", "no-cache")
	r.w.Header().Set("Connection", "keep-alive")
	r.w.WriteHeader(r.code)

	// Create a stream writer that wraps the response writer
	w := &streamWriter{w: r.w}

	// Execute the writer function once
	return r.writer(w)
}

// StreamWriter implementation
type streamWriter struct {
	w http.ResponseWriter
}

func (w *streamWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *streamWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *streamWriter) Flush() error {
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}
