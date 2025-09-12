package modify

import (
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/options"
	supportfile "github.com/goravel/framework/support/file"
)

type FileTestSuite struct {
	suite.Suite
	tempDir string
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}

func (s *FileTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func (s *FileTestSuite) TestOverwrite() {
	tests := []struct {
		name        string
		setup       func() string
		content     string
		force       bool
		expectError bool
		assert      func(path string, err error)
	}{
		{
			name: "overwrite new file without force",
			setup: func() string {
				return filepath.Join(s.tempDir, "new_file.txt")
			},
			content:     "new content",
			force:       false,
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
				content, readErr := supportfile.GetContent(path)
				s.NoError(readErr)
				s.Equal("new content", content)
				s.NoError(supportfile.Remove(path))
			},
		},
		{
			name: "overwrite existing file without force",
			setup: func() string {
				path := filepath.Join(s.tempDir, "existing_file.txt")
				s.NoError(supportfile.PutContent(path, "old content"))
				return path
			},
			content:     "new content",
			force:       false,
			expectError: true,
			assert: func(path string, err error) {
				s.NoError(err)

				// File should not be overwritten
				content, readErr := supportfile.GetContent(path)
				s.NoError(readErr)
				s.Equal("old content", content)
				s.NoError(supportfile.Remove(path))
			},
		},
		{
			name: "overwrite existing file with force",
			setup: func() string {
				path := filepath.Join(s.tempDir, "force_file.txt")
				s.NoError(supportfile.PutContent(path, "old content"))
				return path
			},
			content:     "new content",
			force:       true,
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
				content, readErr := supportfile.GetContent(path)
				s.NoError(readErr)
				s.Equal("new content", content)
				s.NoError(supportfile.Remove(path))
			},
		},
		{
			name: "overwrite with empty content",
			setup: func() string {
				return filepath.Join(s.tempDir, "empty_file.txt")
			},
			content:     "",
			force:       false,
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
				content, readErr := supportfile.GetContent(path)
				s.NoError(readErr)
				s.Empty(content)
				s.NoError(supportfile.Remove(path))
			},
		},
		{
			name: "overwrite with special characters",
			setup: func() string {
				return filepath.Join(s.tempDir, "special_file.txt")
			},
			content:     "content with\nnewlines\tand\ttabs",
			force:       false,
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
				content, readErr := supportfile.GetContent(path)
				s.NoError(readErr)
				s.Equal("content with\nnewlines\tand\ttabs", content)
				s.NoError(supportfile.Remove(path))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := tt.setup()
			overwriteFile := File(path).Overwrite(tt.content)

			err := overwriteFile.Apply(options.Force(tt.force))

			tt.assert(path, err)
		})
	}
}

func (s *FileTestSuite) TestRemove() {
	tests := []struct {
		name        string
		setup       func() string
		expectError bool
		assert      func(path string, err error)
	}{
		{
			name: "remove existing file",
			setup: func() string {
				path := filepath.Join(s.tempDir, "to_remove.txt")
				s.NoError(supportfile.PutContent(path, "content"))
				return path
			},
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
				s.False(supportfile.Exists(path))
				s.NoError(supportfile.Remove(path))
			},
		},
		{
			name: "remove non-existent file",
			setup: func() string {
				return filepath.Join(s.tempDir, "non_existent.txt")
			},
			expectError: false, // RemoveFile doesn't return error for non-existent files
			assert: func(path string, err error) {
				s.NoError(err)
				s.False(supportfile.Exists(path))
			},
		},
		{
			name: "remove empty path",
			setup: func() string {
				return ""
			},
			expectError: false,
			assert: func(path string, err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := tt.setup()
			removeFile := File(path).Remove()

			err := removeFile.Apply()

			tt.assert(path, err)
		})
	}
}

type GoFileTestSuite struct {
	suite.Suite
	file string
}

func TestGoFileTestSuite(t *testing.T) {
	suite.Run(t, new(GoFileTestSuite))
}

func (s *GoFileTestSuite) SetupTest() {
	s.file = filepath.Join(s.T().TempDir(), "test.go")
}

func (s *GoFileTestSuite) TestModifyGoFile() {
	tests := []struct {
		name     string
		setup    func()
		actions  []modify.Action
		matchers []contractsmatch.GoNode
		assert   func(err error)
	}{
		{
			name: "get file content failed",
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "parse file failed",
			setup: func() {
				s.NoError(supportfile.PutContent(s.file, "package main \n invalid go code"))
			},
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "apply modifier failed",
			setup: func() {
				src := `package main
import "fmt"
func main() {
	fmt.Println("Hello, test!")
}
`
				s.Require().NoError(supportfile.PutContent(s.file, src))
			},
			matchers: []contractsmatch.GoNode{
				match.BasicLit("Hello, test!"),
			},
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "apply modifier success",
			setup: func() {
				src := `package main
import "fmt"
func main() {
	fmt.Println("Hello, test!")
}
`
				s.Require().NoError(supportfile.PutContent(s.file, src))
			},
			actions: []modify.Action{func(cursor *dstutil.Cursor) {
				cursor.Replace(&dst.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote("Hello, test!!!"),
				})
			}},
			matchers: []contractsmatch.GoNode{
				match.BasicLit(strconv.Quote("Hello, test!")),
			},
			assert: func(err error) {
				s.NoError(err)
				content, err := supportfile.GetContent(s.file)
				s.NoError(err)
				s.Contains(content, `fmt.Println("Hello, test!!!")`)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}
			tt.assert(GoFile(s.file).Find(tt.matchers).Modify(tt.actions...).Apply())
		})
	}
}

func TestWhenFacade(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		called := false
		apply := &dummyApply{called: &called}
		modifier := WhenFacade("Auth", apply)

		err := modifier.Apply(options.Facade("Auth"))
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("no match", func(t *testing.T) {
		called := false
		apply := &dummyApply{called: &called}
		modifier := WhenFacade("Auth", apply)

		err := modifier.Apply(options.Facade("DB"))
		assert.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("apply error", func(t *testing.T) {
		called := false
		apply := &dummyApply{called: &called, shouldErr: true}
		modifier := WhenFacade("Auth", apply)

		err := modifier.Apply(options.Facade("Auth"))
		assert.Equal(t, assert.AnError, err)
		assert.True(t, called)
	})
}

func TestWhenNoFacades(t *testing.T) {
	t.Run("no facades exist", func(t *testing.T) {
		called := false
		apply := &dummyApply{called: &called}
		modifier := WhenNoFacades([]string{"Auth", "DB"}, apply)
		err := modifier.Apply(options.Facade("Auth"))

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("facade exists", func(t *testing.T) {
		called := false
		apply := &dummyApply{called: &called}
		modifier := WhenNoFacades([]string{"Auth", "DB"}, apply)

		path := facadeToFilepath("DB")
		err := supportfile.PutContent(path, "package facades\n")
		assert.NoError(t, err)

		defer func() {
			assert.NoError(t, supportfile.Remove(path))
		}()

		err = modifier.Apply(options.Facade("Auth"))
		assert.NoError(t, err)
		assert.False(t, called)
	})
}

type dummyApply struct {
	called    *bool
	shouldErr bool
}

func (d *dummyApply) Apply(...modify.Option) error {
	*d.called = true
	if d.shouldErr {
		return assert.AnError
	}
	return nil
}
