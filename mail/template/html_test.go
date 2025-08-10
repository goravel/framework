package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultEngine_Render(t *testing.T) {
	tempDir := t.TempDir()

	templateContent := `Hello {{.Name}}, welcome to {{.App}}!`
	templatePath := filepath.Join(tempDir, "welcome.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	engine := NewHtml(tempDir)

	data := map[string]string{
		"Name": "John",
		"App":  "Goravel",
	}

	result, err := engine.Render("welcome.html", data)
	assert.NoError(t, err)
	assert.Equal(t, "Hello John, welcome to Goravel!", result)
}

func TestDefaultEngine_RenderTemplateNotFound(t *testing.T) {
	tempDir := t.TempDir()
	engine := NewHtml(tempDir)

	_, err := engine.Render("nonexistent.html", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestDefaultEngine_RenderWithCache(t *testing.T) {
	tempDir := t.TempDir()

	templateContent := `Hello {{.Name}}!`
	templatePath := filepath.Join(tempDir, "cached.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	engine := NewHtml(tempDir)
	data := map[string]string{"Name": "Test"}

	// First render - should parse and cache
	result1, err := engine.Render("cached.html", data)
	assert.NoError(t, err)
	assert.Equal(t, "Hello Test!", result1)

	// Second render - should use cache
	result2, err := engine.Render("cached.html", data)
	assert.NoError(t, err)
	assert.Equal(t, "Hello Test!", result2)
	assert.Equal(t, result1, result2)
}

func TestDefaultEngine_RenderError(t *testing.T) {
	tempDir := t.TempDir()

	templateContent := `Hello {{.Name | invalidFunc}}!`
	templatePath := filepath.Join(tempDir, "bad.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	engine := NewHtml(tempDir)

	_, err = engine.Render("bad.html", map[string]string{"Name": "Test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}
