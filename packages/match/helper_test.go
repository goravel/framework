package match

import (
	"go/token"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/packages/match"
)

type MatchHelperTestSuite struct {
	suite.Suite
	source *dst.File
}

func (s *MatchHelperTestSuite) SetupTest() {
	var err error
	s.source, err = decorator.Parse(`package config

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"key": "value",
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
	s.Require().NoError(err)
}

func (s *MatchHelperTestSuite) TearDownTest() {}

func TestHelperTestSuite(t *testing.T) {
	suite.Run(t, new(MatchHelperTestSuite))
}

func (s *MatchHelperTestSuite) match(matchers []match.GoNode) (matched dst.Node) {
	var current int
	dstutil.Apply(s.source, func(cursor *dstutil.Cursor) bool {
		if current >= len(matchers) {
			return false
		}
		if matchers[current].MatchCursor(cursor) {
			current++
			if current == len(matchers) {
				matched = cursor.Node()
				return false
			}
		}

		return true
	}, nil)

	return
}

func (s *MatchHelperTestSuite) TestHelper() {
	tests := []struct {
		name     string
		matchers []match.GoNode
		assert   func(node dst.Node)
	}{
		{
			name:     "match config",
			matchers: Config("app.key"),
			assert: func(node dst.Node) {
				KeyValueExpr(
					BasicLit(strconv.Quote("exist")),
					BasicLit(strconv.Quote("value")),
				)
			},
		},
		{
			name: "match imports",
			matchers: []match.GoNode{
				Imports(),
			},
			assert: func(node dst.Node) {
				n, ok := node.(*dst.GenDecl)
				s.True(ok)
				s.Equal(token.IMPORT, n.Tok)
				s.Len(n.Specs, 4)
			},
		},
		{
			name:     "match providers",
			matchers: Providers(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "foundation"},
						Sel: &dst.Ident{Name: "ServiceProvider"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 2)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			matched := s.match(tt.matchers)
			s.NotNil(matched)
			tt.assert(matched)
		})
	}
}
