package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/support"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/maps"
)

type BodyImpl struct {
	formValues map[string]any
	fileFields map[string]string
}

func NewBody() support.Body {
	return &BodyImpl{
		formValues: make(map[string]any),
		fileFields: make(map[string]string),
	}
}

func (r *BodyImpl) SetFields(fields map[string]any) support.Body {
	r.formValues = collect.Merge(r.formValues, fields)
	return r
}

func (r *BodyImpl) SetField(key string, value any) support.Body {
	maps.Add(r.formValues, key, value)
	return r
}

func (r *BodyImpl) GetField(key string) any {
	return maps.Get(r.formValues, key)
}

func (r *BodyImpl) SetFile(fieldName, filePath string) support.Body {
	r.fileFields[fieldName] = filePath
	return r
}

func (r *BodyImpl) Build() (io.Reader, error) {
	if len(r.fileFields) > 0 {
		return r.buildMultipartBody()
	}
	return r.buildJSONBody()
}

func (r *BodyImpl) buildMultipartBody() (io.Reader, error) {
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

	return payload, nil
}

func (r *BodyImpl) addFormFields(writer *multipart.Writer) error {
	for key, val := range r.formValues {
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

func (r *BodyImpl) buildJSONBody() (io.Reader, error) {
	jsonBody, err := json.Marshal(r.formValues)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jsonBody), nil
}
