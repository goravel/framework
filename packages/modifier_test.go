package packages

import (
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/file"
)

type ModifyGoFileTestSuite struct {
	suite.Suite
	dir  string
	file string
}

func (s *ModifyGoFileTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
	s.file = "test.go"
}

func (s *ModifyGoFileTestSuite) TearDownTest() {}

func TestModifyGoFileTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyGoFileTestSuite))
}

func (s *ModifyGoFileTestSuite) TestModifyGoFile() {
	cases := []struct {
		name   string
		setup  func(g *ModifyGoFile)
		assert func(err error)
	}{
		{
			name: "get file content failed",
			setup: func(g *ModifyGoFile) {
				g.File = s.file
			},
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "parse file failed",
			setup: func(g *ModifyGoFile) {
				g.File = s.file
				s.NoError(file.PutContent(filepath.Join(s.dir, s.file), "package main \n invalid go code"))
			},
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "apply modifier failed",
			setup: func(g *ModifyGoFile) {
				g.File = s.file
				src := `package main
import "fmt"
func main() {
	fmt.Println("Hello, test!")
}
`
				s.NoError(file.PutContent(filepath.Join(s.dir, s.file), src))
				g.Modifiers = []packages.GoNodeModifier{
					&ModifyGoNode{
						Action: func(_ *dstutil.Cursor) {

						},
						Matchers: []packages.GoNodeMatcher{
							MatchBasicLit("Hello, test!"),
						},
					},
				}
			},
			assert: func(err error) {
				s.Error(err)
			},
		},
		{
			name: "apply modifier success",
			setup: func(g *ModifyGoFile) {
				g.File = s.file
				src := `package main
import "fmt"
func main() {
	fmt.Println("Hello, test!")
}
`
				s.NoError(file.PutContent(filepath.Join(s.dir, s.file), src))
				g.Modifiers = []packages.GoNodeModifier{
					&ModifyGoNode{
						Action: func(cursor *dstutil.Cursor) {
							cursor.Replace(&dst.BasicLit{
								Kind:  token.STRING,
								Value: strconv.Quote("Hello, test!!!"),
							})
						},
						Matchers: []packages.GoNodeMatcher{
							MatchBasicLit(strconv.Quote("Hello, test!")),
						},
					},
				}
			},
			assert: func(err error) {
				s.NoError(err)
				content, err := file.GetContent(filepath.Join(s.dir, s.file))
				s.NoError(err)
				s.Contains(content, `fmt.Println("Hello, test!!!")`)
			},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			g := &ModifyGoFile{}
			tc.setup(g)
			tc.assert(g.Apply(s.dir))
		})
	}
}
