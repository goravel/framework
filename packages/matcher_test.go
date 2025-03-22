package packages

import (
	"go/token"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/packages"
)

type MatchGoNodeTestSuite struct {
	suite.Suite
	source *dst.File
}

func (s *MatchGoNodeTestSuite) SetupTest() {
	var err error
	s.source, err = decorator.Parse(`package config

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name": config.Env("APP_NAME", "Goravel"),
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
		},
	})
}`)
	s.Require().NoError(err)
}

func (s *MatchGoNodeTestSuite) TearDownTest() {}

func TestMatchGoNodeTestSuite(t *testing.T) {
	suite.Run(t, new(MatchGoNodeTestSuite))
}

func (s *MatchGoNodeTestSuite) match(matcher packages.GoNodeMatcher) (matched dst.Node) {
	dst.Inspect(s.source, func(node dst.Node) bool {
		if matcher.MatchNode(node) {
			matched = node
			return false
		}
		return true
	})

	return
}

func (s *MatchGoNodeTestSuite) TestMatch() {

	cases := []struct {
		name    string
		matcher packages.GoNodeMatcher
		assert  func(node dst.Node)
	}{
		{
			name: "match array type",
			matcher: MatchArrayType(
				MatchSelectorExpr(
					MatchIdent("foundation"),
					MatchIdent("ServiceProvider"),
				),
				MatchAnyNode(),
			),
			assert: func(node dst.Node) {
				s.True(
					MatchEqualNode(node).MatchNode(
						&dst.ArrayType{
							Elt: &dst.SelectorExpr{
								X:   &dst.Ident{Name: "foundation"},
								Sel: &dst.Ident{Name: "ServiceProvider"},
							},
						},
					),
				)
			},
		},
		{
			name: "match  call expr",
			matcher: MatchCallExpr(
				MatchSelectorExpr(
					MatchIdent("config"),
					MatchIdent("Add"),
				),
				MatchAnyNodes(),
			),
			assert: func(node dst.Node) {
				call, ok := node.(*dst.CallExpr)
				s.True(ok)

				// check if the function is config.Add
				s.True(
					MatchEqualNode(call.Fun).MatchNode(
						&dst.SelectorExpr{
							X:   &dst.Ident{Name: "config"},
							Sel: &dst.Ident{Name: "Add"},
						},
					),
				)

				// check if the first argument is "app"
				s.True(
					MatchEqualNode(call.Args[0]).MatchNode(
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: strconv.Quote("app"),
						},
					),
				)
			},
		},
		{
			name:    "match composite lit",
			matcher: MatchCompositeLit(MatchTypeOf(&dst.MapType{})),
			assert: func(node dst.Node) {
				s.False(
					MatchEqualNode(node).MatchNode(
						&dst.CompositeLit{
							Type: &dst.MapType{
								Key:   &dst.Ident{Name: "string"},
								Value: &dst.Ident{Name: "any"},
							},
							Elts: []dst.Expr{
								&dst.KeyValueExpr{},
								&dst.KeyValueExpr{},
							},
						},
					),
				)

				// check if the type is map[string]any
				cl, ok := node.(*dst.CompositeLit)
				s.True(ok)
				s.True(
					MatchEqualNode(cl.Type).MatchNode(&dst.MapType{
						Key:   &dst.Ident{Name: "string"},
						Value: &dst.Ident{Name: "any"},
					}),
				)

				// check if the elements are 2
				s.Len(cl.Elts, 2)
			},
		},
		{
			name:    "match function",
			matcher: MatchFuncDecl(MatchIdent("Boot")),
			assert: func(node dst.Node) {
				fn, ok := node.(*dst.FuncDecl)
				s.True(ok)
				s.Equal(fn.Name.Name, "Boot")
			},
		},
		{
			name:    "match import spec",
			matcher: MatchImportSpec("github.com/goravel/framework/facades"),
			assert: func(node dst.Node) {
				s.True(
					MatchEqualNode(node).MatchNode(
						&dst.ImportSpec{
							Path: &dst.BasicLit{
								Kind:  token.STRING,
								Value: strconv.Quote("github.com/goravel/framework/facades"),
							},
						},
					),
				)
			},
		},
		{
			name:    "match key value expr",
			matcher: MatchKeyValueExpr(MatchBasicLit(strconv.Quote("providers")), MatchAnyNode()),
			assert: func(node dst.Node) {
				kv, ok := node.(*dst.KeyValueExpr)
				s.True(ok)
				s.True(
					MatchEqualNode(kv.Key).MatchNode(
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: strconv.Quote("providers"),
						},
					),
				)
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			matched := s.match(tc.matcher)
			s.NotNil(matched)
			tc.assert(matched)
		})
	}

}
