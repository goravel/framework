package file

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/testing/file"
)

func TestClientOriginalExtension(t *testing.T) {
	assert.Equal(t, ClientOriginalExtension("logo.png"), "png")
}

func TestContain(t *testing.T) {
	assert.True(t, Contain("../constant.go", "Version"))
}

func TestContains(t *testing.T) {
	t.Run("file not exists", func(t *testing.T) {
		assert.False(t, Contains("nonexistent.go", "content"))
	})

	t.Run("file exists and contains search string", func(t *testing.T) {
		assert.True(t, Contains("../constant.go", "Version"))
	})

	t.Run("file exists but does not contain search string", func(t *testing.T) {
		assert.False(t, Contains("../constant.go", "NonExistentString123"))
	})

	t.Run("normalize line endings - LF file with LF search", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_lf.txt")
		content := "line1\nline2\nline3"
		assert.NoError(t, PutContent(filePath, content))
		assert.True(t, Contains(filePath, "line1\nline2"))
	})

	t.Run("normalize line endings - CRLF file with LF search", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_crlf.txt")
		// Simulate Windows CRLF line endings
		content := "line1\r\nline2\r\nline3"
		assert.NoError(t, PutContent(filePath, content))
		// Search with LF should still work due to normalization
		assert.True(t, Contains(filePath, "line1\nline2"))
	})

	t.Run("normalize line endings - CRLF file with CRLF search", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_crlf2.txt")
		content := "line1\r\nline2\r\nline3"
		assert.NoError(t, PutContent(filePath, content))
		// Search with CRLF should also work
		assert.True(t, Contains(filePath, "line1\r\nline2"))
	})

	t.Run("normalize line endings - LF file with CRLF search", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_lf2.txt")
		content := "line1\nline2\nline3"
		assert.NoError(t, PutContent(filePath, content))
		// Search with CRLF should match due to normalization
		assert.True(t, Contains(filePath, "line1\r\nline2"))
	})

	t.Run("multiline content with mixed line endings", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_mixed.txt")
		// Mixed line endings (some LF, some CRLF)
		content := "line1\nline2\r\nline3\nline4"
		assert.NoError(t, PutContent(filePath, content))
		// After normalization, all should be LF
		assert.True(t, Contains(filePath, "line2\nline3"))
	})

	t.Run("search for code with line breaks", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "test_code.go")
		content := "package main\r\n\r\nfunc main() {\r\n\tfmt.Println(\"hello\")\r\n}"
		assert.NoError(t, PutContent(filePath, content))
		// Search for code snippet with LF
		assert.True(t, Contains(filePath, "func main() {\n\tfmt.Println(\"hello\")"))
	})

	t.Run("empty file", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "empty.txt")
		assert.NoError(t, PutContent(filePath, ""))
		assert.False(t, Contains(filePath, "anything"))
		assert.True(t, Contains(filePath, ""))
	})

	t.Run("search string with no line breaks", func(t *testing.T) {
		filePath := path.Join(t.TempDir(), "simple.txt")
		assert.NoError(t, PutContent(filePath, "hello world"))
		assert.True(t, Contains(filePath, "hello"))
		assert.True(t, Contains(filePath, "world"))
		assert.True(t, Contains(filePath, "hello world"))
	})
}

func TestCopyFile(t *testing.T) {
	t.Run("copy file successfully", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), ".env.example")
		dst := filepath.Join(t.TempDir(), ".env")
		content := "example env content"

		assert.NoError(t, os.WriteFile(src, []byte(content), os.ModePerm))
		assert.True(t, Exists(src))

		assert.NoError(t, Copy(src, dst))
		assert.True(t, Exists(dst))

		// Verify content was copied correctly
		dstContent, err := GetContent(dst)
		assert.NoError(t, err)
		assert.Equal(t, content, dstContent)
	})

	t.Run("source file does not exist", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "nonexistent.txt")
		dst := filepath.Join(t.TempDir(), "destination.txt")

		assert.Error(t, Copy(src, dst))
		assert.False(t, Exists(dst))
	})

	t.Run("copy to existing file overwrites", func(t *testing.T) {
		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "source.txt")
		dst := filepath.Join(tmpDir, "destination.txt")
		srcContent := "new content"
		oldContent := "old content"

		// Create destination with old content
		assert.NoError(t, PutContent(dst, oldContent))
		// Create source with new content
		assert.NoError(t, PutContent(src, srcContent))

		// Copy should overwrite
		assert.NoError(t, Copy(src, dst))

		result, err := GetContent(dst)
		assert.NoError(t, err)
		assert.Equal(t, srcContent, result)
	})
}

func TestCreate(t *testing.T) {
	filePath := path.Join(t.TempDir(), "goravel.txt")
	assert.Nil(t, PutContent(filePath, `goravel`))
	assert.Equal(t, 1, file.GetLineNum(filePath))
	assert.True(t, Exists(filePath))
}

func TestExists(t *testing.T) {
	assert.True(t, Exists("file.go"))
}

func TestExtension(t *testing.T) {
	extension, err := Extension("file.go")
	assert.Nil(t, err)
	assert.Equal(t, "txt", extension)
}

func TestLastModified(t *testing.T) {
	ti, err := LastModified("../../logo.png", "UTC")
	assert.Nil(t, err)
	assert.NotNil(t, ti)
}

func TestMimeType(t *testing.T) {
	mimeType, err := MimeType("../../logo.png")
	assert.Nil(t, err)
	assert.Equal(t, "image/png", mimeType)
}

func TestRemove(t *testing.T) {
	pwd, _ := os.Getwd()
	filePath := path.Join(pwd, "goravel/goravel.txt")
	assert.Nil(t, PutContent(filePath, `goravel`))

	assert.Nil(t, Remove(filePath))
	assert.Nil(t, Remove(path.Join(pwd, "goravel")))
}

func TestSize(t *testing.T) {
	size, err := Size("../../logo.png")
	assert.Nil(t, err)
	assert.Equal(t, int64(10853), size)
}

func TestGetContent(t *testing.T) {
	// file not exists
	content, err := GetContent("files.go")
	assert.NotNil(t, err)
	assert.Empty(t, content)
	// get content successfully
	content, err = GetContent("../constant.go")
	assert.Nil(t, err)
	assert.NotNil(t, content)
	assert.Contains(t, content, "Version")
}

func TestPutContent(t *testing.T) {
	if !env.IsWindows() {
		// directory creation failure
		assert.Error(t, PutContent("/proc/invalid/file.txt", "content"))
		// write failure (create read-only dir)
		readOnlyDir := path.Join(t.TempDir(), "readonly")
		assert.NoError(t, os.Mkdir(readOnlyDir, 0444))
		assert.Error(t, PutContent(path.Join(readOnlyDir, "file.txt"), "content"))
	}
	// create a file and put content
	filePath := path.Join(t.TempDir(), "goravel.txt")
	assert.NoError(t, PutContent(filePath, "goravel"))
	assert.True(t, Contain(filePath, "goravel"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
	// append content
	assert.NoError(t, PutContent(filePath, "\nframework", WithAppend()))
	assert.True(t, Contain(filePath, "goravel\nframework"))
	assert.Equal(t, 2, file.GetLineNum(filePath))
	// overwrite content
	assert.NoError(t, PutContent(filePath, "welcome", WithMode(0644)))
	assert.False(t, Contain(filePath, "goravel\nframework"))
	assert.True(t, Contain(filePath, "welcome"))
	assert.Equal(t, 1, file.GetLineNum(filePath))
}
