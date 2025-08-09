package template

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"sync"
)

type DefaultEngine struct {
	viewsPath string
	cache     sync.Map
}

func NewDefaultEngine(viewsPath string) *DefaultEngine {
	return &DefaultEngine{
		viewsPath: viewsPath,
	}
}

func (r *DefaultEngine) Render(path string, data any) (string, error) {
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

func (r *DefaultEngine) getTemplate(templatePath string) (*template.Template, error) {
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
