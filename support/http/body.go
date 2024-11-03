package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/support"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

type BodyType int

const (
	BodyTypeJSON BodyType = iota
	BodyTypeForm
	BodyTypeMultipart
)

type BodyImpl struct {
	data       map[string]any
	fileFields map[string]string
	bodyType   BodyType
}

// NewBody creates a new BodyImpl instance with an optional bodyType argument.
// If bodyType is not provided, it defaults to BodyTypeJSON.
func NewBody(bodyType ...BodyType) support.Body {
	bt := BodyTypeJSON
	if len(bodyType) > 0 {
		bt = bodyType[0]
	}

	return &BodyImpl{
		data:       make(map[string]any),
		fileFields: make(map[string]string),
		bodyType:   bt,
	}
}

func (r *BodyImpl) SetFields(fields map[string]any) support.Body {
	r.data = collect.Merge(r.data, fields)
	return r
}

func (r *BodyImpl) SetField(key string, value any) support.Body {
	maps.Add(r.data, key, value)
	return r
}

func (r *BodyImpl) GetField(key string) any {
	return maps.Get(r.data, key)
}

func (r *BodyImpl) SetFiles(files map[string]string) support.Body {
	r.fileFields = collect.Merge(r.fileFields, files)
	r.bodyType = BodyTypeMultipart
	return r
}

func (r *BodyImpl) SetFile(fieldName, filePath string) support.Body {
	r.fileFields[fieldName] = filePath
	r.bodyType = BodyTypeMultipart
	return r
}

func (r *BodyImpl) Build() (support.BodyContent, error) {
	switch r.bodyType {
	case BodyTypeMultipart:
		return r.buildMultipartBody()
	case BodyTypeForm:
		return r.buildFormBody()
	default:
		return r.buildJSONBody()
	}

}

func (r *BodyImpl) buildMultipartBody() (support.BodyContent, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if err := r.addFormFields(writer); err != nil {
		return nil, err
	}

	if err := r.addFiles(writer); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return newBodyContent(payload, writer.FormDataContentType()), nil
}

func (r *BodyImpl) buildFormBody() (support.BodyContent, error) {
	formData := url.Values{}
	for key, val := range r.data {
		formData.Add(key, cast.ToString(val))
	}

	return newBodyContent(strings.NewReader(formData.Encode()), ContentTypeFormURLEncoded), nil
}

func (r *BodyImpl) buildJSONBody() (support.BodyContent, error) {
	jsonBody, err := json.Marshal(r.data)
	if err != nil {
		return nil, err
	}

	return newBodyContent(bytes.NewReader(jsonBody), ContentTypeJSON), nil
}

func (r *BodyImpl) addFormFields(writer *multipart.Writer) error {
	for key, val := range r.data {
		if err := writer.WriteField(key, cast.ToString(val)); err != nil {
			return err
		}
	}
	return nil
}

func (r *BodyImpl) addFiles(writer *multipart.Writer) error {
	for fieldName, filePath := range r.fileFields {
		if err := r.addFile(writer, fieldName, filePath); err != nil {
			return err
		}
	}
	return nil
}

func (r *BodyImpl) addFile(writer *multipart.Writer, fieldName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(file.Name()))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	return err
}
