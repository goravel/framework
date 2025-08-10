package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestHtml_Render(t *testing.T) {
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

func TestHtml_RenderTemplateNotFound(t *testing.T) {
	tempDir := t.TempDir()
	engine := NewHtml(tempDir)

	_, err := engine.Render("nonexistent.html", nil)
	assert.ErrorIs(t, err, errors.MailTemplateParseFailed)
}

func TestHtml_RenderWithCache(t *testing.T) {
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

func TestHtml_RenderExecutionError(t *testing.T) {
	tempDir := t.TempDir()

	// Create template that parses successfully but fails during execution
	// This will fail when trying to access a field on nil
	templateContent := `Hello {{.User.Name}}!`
	templatePath := filepath.Join(tempDir, "bad.html")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	engine := NewHtml(tempDir)

	// Pass data where .User is nil, causing execution error
	_, err = engine.Render("bad.html", map[string]any{"User": nil})
	assert.ErrorIs(t, err, errors.MailTemplateExecutionFailed)
}
