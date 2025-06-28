package modify

import (
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/stretchr/testify/suite"

	contractsmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	supportfile "github.com/goravel/framework/support/file"
)

type ModifyGoFileTestSuite struct {
	suite.Suite
	file string
}

func (s *ModifyGoFileTestSuite) SetupTest() {
	s.file = filepath.Join(s.T().TempDir(), "test.go")
}

func (s *ModifyGoFileTestSuite) TearDownTest() {}

func TestModifyGoFileTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyGoFileTestSuite))
}

func (s *ModifyGoFileTestSuite) TestModifyGoFile() {
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
