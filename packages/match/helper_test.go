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
	configChained   *dst.File
	configVariable  *dst.File
	console         *dst.File
	database        *dst.File
	jobs            *dst.File
	serviceProvider *dst.File
	validation      *dst.File
}

func (s *MatchHelperTestSuite) SetupTest() {
	var err error
	s.configChained, err = decorator.Parse(`package config

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	facades.Config().Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"key": "value",
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
	s.Require().NoError(err)
	s.configVariable, err = decorator.Parse(`package config

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
	s.console, err = decorator.Parse(`package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
	"goravel/app/console/commands"
)

type Kernel struct {
}

func (kernel Kernel) Schedule() []schedule.Event {
	return []schedule.Event{}
}

func (kernel Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.Test{},
	}
}`)
	s.Require().NoError(err)
	s.database, err = decorator.Parse(`package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/migrations"
	"goravel/database/seeders"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20240915060148CreateUsersTable{},
		&migrations.M20250301000000CreateFailedJobsTable{},
	}
}

func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}`)
	s.Require().NoError(err)
	s.validation, err = decorator.Parse(`package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"

	"goravel/app/filters"
	"goravel/app/rules"
)

type ValidationServiceProvider struct {
}

func (receiver *ValidationServiceProvider) Register(app foundation.Application) {

}

func (receiver *ValidationServiceProvider) Boot(app foundation.Application) {
	if err := facades.Validation().AddRules(receiver.rules()); err != nil {
		facades.Log().Errorf("add rules error: %+v", err)
	}
	if err := facades.Validation().AddFilters(receiver.filters()); err != nil {
		facades.Log().Errorf("add filters error: %+v", err)
	}
}

func (receiver *ValidationServiceProvider) rules() []validation.Rule {
	return []validation.Rule{
		&rules.Uppercase{},
	}
}

func (receiver *ValidationServiceProvider) filters() []validation.Filter {
	return []validation.Filter{
		&filters.ToInt{},
	}
}`)
	s.Require().NoError(err)
	s.jobs, err = decorator.Parse(`package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	
	"goravel/app/jobs"
)

type QueueServiceProvider struct {
}

func (receiver *QueueServiceProvider) Register(app foundation.Application) {
	facades.Queue().Register(receiver.Jobs())
}

func (receiver *QueueServiceProvider) Boot(app foundation.Application) {

}

func (receiver *QueueServiceProvider) Jobs() []queue.Job {
	return []queue.Job{
		&jobs.Test{},
	}
}
`)
	s.Require().NoError(err)

	s.serviceProvider, err = decorator.Parse(`package providers

import (
	"github.com/goravel/framework/contracts/foundation"
)

type AppServiceProvider struct {
}

func (receiver *AppServiceProvider) Register(app foundation.Application) {

}

func (receiver *AppServiceProvider) Boot(app foundation.Application) {

}
`)
	s.Require().NoError(err)
}

func (s *MatchHelperTestSuite) TearDownTest() {}

func TestMatchHelperTestSuite(t *testing.T) {
	suite.Run(t, new(MatchHelperTestSuite))
}

func (s *MatchHelperTestSuite) TestHelper() {
	tests := []struct {
		name     string
		file     *dst.File
		matchers []match.GoNode
		assert   func(node dst.Node)
	}{
		{
			name:     "match config (chained call)",
			file:     s.configChained,
			matchers: Config("app.key"),
			assert: func(node dst.Node) {
				s.True(KeyValueExpr(
					BasicLit(strconv.Quote("key")),
					BasicLit(strconv.Quote("value")),
				).MatchNode(node))
			},
		},
		{
			name:     "match config (variable-based call)",
			file:     s.configVariable,
			matchers: Config("app.key"),
			assert: func(node dst.Node) {
				s.True(KeyValueExpr(
					BasicLit(strconv.Quote("key")),
					BasicLit(strconv.Quote("value")),
				).MatchNode(node))
			},
		},
		{
			name:     "match imports",
			file:     s.configVariable,
			matchers: Imports(),
			assert: func(node dst.Node) {
				n, ok := node.(*dst.GenDecl)
				s.True(ok)
				s.Equal(token.IMPORT, n.Tok)
				s.Len(n.Specs, 4)
			},
		},
		{
			name:     "match providers",
			file:     s.configVariable,
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
		{
			name:     "match migrations",
			file:     s.database,
			matchers: Migrations(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "schema"},
						Sel: &dst.Ident{Name: "Migration"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 2)
			},
		},
		{
			name:     "match seeders",
			file:     s.database,
			matchers: Seeders(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "seeder"},
						Sel: &dst.Ident{Name: "Seeder"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 1)
			},
		},
		{
			name:     "match commands",
			file:     s.console,
			matchers: Commands(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "console"},
						Sel: &dst.Ident{Name: "Command"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 1)
			},
		},
		{
			name:     "match validation rules",
			file:     s.validation,
			matchers: ValidationRules(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "validation"},
						Sel: &dst.Ident{Name: "Rule"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 1)
			},
		},
		{
			name:     "match validation filters",
			file:     s.validation,
			matchers: ValidationFilters(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "validation"},
						Sel: &dst.Ident{Name: "Filter"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 1)
			},
		},
		{
			name:     "match jobs",
			file:     s.jobs,
			matchers: Jobs(),
			assert: func(node dst.Node) {
				s.True(CompositeLit(EqualNode(&dst.ArrayType{
					Elt: &dst.SelectorExpr{
						X:   &dst.Ident{Name: "queue"},
						Sel: &dst.Ident{Name: "Job"},
					},
				})).MatchNode(node))
				s.Len(node.(*dst.CompositeLit).Elts, 1)
			},
		},
		{
			name:     "match servicer provider's register function",
			file:     s.serviceProvider,
			matchers: RegisterFunc(),
			assert: func(node dst.Node) {
				fn, ok := node.(*dst.FuncDecl)
				s.True(ok)
				s.True(fn.Name.Name == "Register")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			matched := s.match(tt.file, tt.matchers)
			s.NotNil(matched)
			tt.assert(matched)
		})
	}
}

func (s *MatchHelperTestSuite) match(source *dst.File, matchers []match.GoNode) (matched dst.Node) {
	var current int
	dstutil.Apply(source, func(cursor *dstutil.Cursor) bool {
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
