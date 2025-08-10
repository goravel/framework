package template

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"sync"
)

type Html struct {
	viewsPath string
	cache     sync.Map
}

func NewHtml(viewsPath string) *Html {
	return &Html{
		viewsPath: viewsPath,
	}
}

func (r *Html) Render(path string, data any) (string, error) {
	templatePath := filepath.Join(r.viewsPath, path)
	tmpl, err := r.getTemplate(templatePath)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", path, err)
	}

	return buf.String(), nil
}

func (r *Html) getTemplate(templatePath string) (*template.Template, error) {
	if cached, ok := r.cache.Load(templatePath); ok {
		return cached.(*template.Template), nil
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	r.cache.LoadOrStore(templatePath, tmpl)
	return tmpl, nil
}
