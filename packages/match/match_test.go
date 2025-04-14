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
	facades "github.com/goravel/framework/facades"
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

func (s *MatchGoNodeTestSuite) match(matcher match.GoNode) (matched dst.Node) {
	dstutil.Apply(s.source, func(cursor *dstutil.Cursor) bool {
		if matcher.MatchCursor(cursor) {
			matched = cursor.Node()
			return false
		}
		return true
	}, nil)

	return
}

func (s *MatchGoNodeTestSuite) TestMatch() {

	tests := []struct {
		name    string
		matcher match.GoNode
		assert  func(node dst.Node)
	}{
		{
			name: "match array type",
			matcher: ArrayType(
				SelectorExpr(
					Ident("foundation"),
					Ident("ServiceProvider"),
				),
				AnyNode(),
			),
			assert: func(node dst.Node) {
				s.True(
					EqualNode(node).MatchNode(
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
			matcher: CallExpr(
				SelectorExpr(
					Ident("config"),
					Ident("Add"),
				),
				AnyNodes(),
			),
			assert: func(node dst.Node) {
				call, ok := node.(*dst.CallExpr)
				s.True(ok)

				// check if the function is config.Add
				s.True(
					EqualNode(call.Fun).MatchNode(
						&dst.SelectorExpr{
							X:   &dst.Ident{Name: "config"},
							Sel: &dst.Ident{Name: "Add"},
						},
					),
				)

				// check if the first argument is "app"
				s.True(
					EqualNode(call.Args[0]).MatchNode(
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
			matcher: CompositeLit(TypeOf(&dst.MapType{})),
			assert: func(node dst.Node) {
				s.False(
					EqualNode(node).MatchNode(
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
					EqualNode(cl.Type).MatchNode(&dst.MapType{
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
			matcher: FuncDecl(Ident("Boot")),
			assert: func(node dst.Node) {
				fn, ok := node.(*dst.FuncDecl)
				s.True(ok)
				s.Equal(fn.Name.Name, "Boot")
			},
		},
		{
			name:    "match import spec",
			matcher: ImportSpec("github.com/goravel/framework/facades", "facades"),
			assert: func(node dst.Node) {
				s.True(
					EqualNode(node).MatchNode(
						&dst.ImportSpec{
							Name: &dst.Ident{Name: "facades"},
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
			name:    "match first of import spec",
			matcher: FirstOf(TypeOf(&dst.ImportSpec{})),
			assert: func(node dst.Node) {
				n, ok := node.(*dst.ImportSpec)
				s.True(ok)
				s.True(ImportSpec("github.com/goravel/framework/auth").MatchNode(n))
			},
		},
		{
			name:    "match first of import spec",
			matcher: LastOf(TypeOf(&dst.ImportSpec{})),
			assert: func(node dst.Node) {
				n, ok := node.(*dst.ImportSpec)
				s.True(ok)
				s.True(ImportSpec("github.com/goravel/framework/facades", "facades").MatchNode(n))
			},
		},
		{
			name:    "match key value expr",
			matcher: KeyValueExpr(BasicLit(strconv.Quote("providers")), AnyNode()),
			assert: func(node dst.Node) {
				kv, ok := node.(*dst.KeyValueExpr)
				s.True(ok)
				s.True(
					EqualNode(kv.Key).MatchNode(
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: strconv.Quote("providers"),
						},
					),
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			matched := s.match(tt.matcher)
			s.NotNil(matched)
			tt.assert(matched)
		})
	}
}
