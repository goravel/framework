package gorm

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/schema"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

type relationUser struct {
	ID       uint
	Name     string
	Address  *relationAddress
	Books    []*relationBook
	Comments []*relationComment
	Phones   []*relationPhone `gorm:"polymorphic:Phoneable"`
	Roles    []*relationRole  `gorm:"many2many:relation_role_users"`
}

type relationAddress struct {
	ID             uint
	RelationUserID uint
	City           string
}

type relationBook struct {
	ID             uint
	RelationUserID uint
	Title          string
	Published      bool
	RelationUser   *relationUser
	Authors        []*relationAuthor
}

type relationAuthor struct {
	ID             uint
	RelationBookID uint
	Name           string
}

type relationComment struct {
	ID             uint
	RelationUserID uint
	DeletedAt      gormio.DeletedAt
}

type relationPhone struct {
	ID            uint
	PhoneableID   uint
	PhoneableType string
	Number        string
}

type relationRole struct {
	ID   uint
	Name string
}

type relationScopedBook struct {
	ID             uint
	RelationUserID uint
}

func (r *relationScopedBook) GlobalScopes() map[string]func(contractsorm.Query) contractsorm.Query {
	return map[string]func(contractsorm.Query) contractsorm.Query{
		"published": func(query contractsorm.Query) contractsorm.Query {
			return query.Where("published", true)
		},
	}
}

type relationScopedUser struct {
	ID    uint
	Books []*relationScopedBook `gorm:"foreignKey:RelationUserID"`
}

type relationOrScopedBook struct {
	ID             uint
	RelationUserID uint
}

func (r *relationOrScopedBook) GlobalScopes() map[string]func(contractsorm.Query) contractsorm.Query {
	return map[string]func(contractsorm.Query) contractsorm.Query{
		"visible": func(query contractsorm.Query) contractsorm.Query {
			return query.Where("published", true).OrWhere("featured", true)
		},
	}
}

type relationOrScopedUser struct {
	ID    uint
	Books []*relationOrScopedBook `gorm:"foreignKey:RelationUserID"`
}

// relationMixedNode is a tree on a table with upper case letters.
type relationMixedNode struct {
	ID       uint
	ParentID *uint
	Children []*relationMixedNode `gorm:"foreignKey:ParentID"`
}

func (r *relationMixedNode) TableName() string {
	return "Nodes"
}

// relationNode is a tree whose relations are on its own table.
type relationNode struct {
	ID        uint
	ParentID  *uint
	Parent    *relationNode   `gorm:"foreignKey:ParentID"`
	Children  []*relationNode `gorm:"foreignKey:ParentID"`
	Friends   []*relationNode `gorm:"many2many:relation_node_friends"`
	DeletedAt gormio.DeletedAt
}

type relationConnectionBook struct {
	ID             uint
	RelationUserID uint
}

func (r *relationConnectionBook) Connection() string {
	return "other"
}

type relationConnectionUser struct {
	ID    uint
	Books []*relationConnectionBook `gorm:"foreignKey:RelationUserID"`
}

func TestRelationExistence(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	require.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil)
	}
	users := func() contractsorm.Query {
		return newQuery().Model(&relationUser{})
	}
	published := func(query contractsorm.Query) contractsorm.Query {
		return query.Where("published", true)
	}

	tests := []struct {
		name        string
		query       func() contractsorm.Query
		dest        any
		expectSql   string
		expectVars  []any
		expectError error
	}{
		{
			name:      "Has with has many",
			query:     func() contractsorm.Query { return users().Has("Books") },
			expectSql: "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name:      "Has with has one",
			query:     func() contractsorm.Query { return users().Has("Address") },
			expectSql: "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_addresses` WHERE `relation_addresses`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name:      "Has with belongs to",
			query:     func() contractsorm.Query { return newQuery().Model(&relationBook{}).Has("RelationUser") },
			expectSql: "SELECT * FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_users` WHERE `relation_users`.`id` = `relation_books`.`relation_user_id`)",
		},
		{
			name:       "Has with polymorphic has many",
			query:      func() contractsorm.Query { return users().Has("Phones") },
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_phones` WHERE `relation_phones`.`phoneable_type` = ? AND `relation_phones`.`phoneable_id` = `relation_users`.`id`)",
			expectVars: []any{"relation_users"},
		},
		{
			name:      "Has with many to many",
			query:     func() contractsorm.Query { return users().Has("Roles") },
			expectSql: "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_roles` INNER JOIN `relation_role_users` ON `relation_role_users`.`relation_role_id` = `relation_roles`.`id` WHERE `relation_role_users`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name:      "Has skips soft deleted related records",
			query:     func() contractsorm.Query { return users().Has("Comments") },
			expectSql: "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_comments` WHERE `relation_comments`.`relation_user_id` = `relation_users`.`id` AND `relation_comments`.`deleted_at` IS NULL)",
		},
		{
			name:       "Has with a count",
			query:      func() contractsorm.Query { return users().Has("Books", ">=", 3) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) >= ?",
			expectVars: []any{int64(3)},
		},
		{
			name:       "Has with an unsigned count",
			query:      func() contractsorm.Query { return users().Has("Books", "=", uint(2)) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) = ?",
			expectVars: []any{int64(2)},
		},
		{
			name:      "Has with a nested relation",
			query:     func() contractsorm.Query { return users().Has("Books.Authors") },
			expectSql: "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_authors` WHERE `relation_authors`.`relation_book_id` = `relation_books`.`id`) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name:       "Has with a nested relation and a count",
			query:      func() contractsorm.Query { return users().Has("Books.Authors", ">", 1) },
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE (SELECT COUNT(*) FROM `relation_authors` WHERE `relation_authors`.`relation_book_id` = `relation_books`.`id`) > ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{int64(1)},
		},
		{
			name:       "Has uses the destination as the model",
			query:      func() contractsorm.Query { return newQuery().Where("name", "goravel").Has("Books") },
			dest:       &[]relationUser{},
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? AND EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name:       "OrHas",
			query:      func() contractsorm.Query { return users().Where("name", "goravel").OrHas("Books") },
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? OR EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name: "Has inside a group uses the model of the query",
			query: func() contractsorm.Query {
				return users().Where("name", "goravel").Where(func(query contractsorm.Query) contractsorm.Query {
					return query.Has("Roles").OrDoesntHave("Books")
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? AND (EXISTS (SELECT 1 FROM `relation_roles` INNER JOIN `relation_role_users` ON `relation_role_users`.`relation_role_id` = `relation_roles`.`id` WHERE `relation_role_users`.`relation_user_id` = `relation_users`.`id`) OR NOT EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`))",
			expectVars: []any{"goravel"},
		},
		{
			name:      "DoesntHave",
			query:     func() contractsorm.Query { return users().DoesntHave("Books") },
			expectSql: "SELECT * FROM `relation_users` WHERE NOT EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name:       "OrDoesntHave",
			query:      func() contractsorm.Query { return users().Where("name", "goravel").OrDoesntHave("Books") },
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? OR NOT EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name:       "WhereHas",
			query:      func() contractsorm.Query { return users().WhereHas("Books", published) },
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{true},
		},
		{
			name:       "WhereHas with a count",
			query:      func() contractsorm.Query { return users().WhereHas("Books", published, "<", 2) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`) < ?",
			expectVars: []any{true, int64(2)},
		},
		{
			name: "WhereHas with a nested relation",
			query: func() contractsorm.Query {
				return users().WhereHas("Books.Authors", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "goravel")
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_authors` WHERE `name` = ? AND `relation_authors`.`relation_book_id` = `relation_books`.`id`) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name: "WhereHas with a callback that uses WhereHas",
			query: func() contractsorm.Query {
				return users().WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
					return query.WhereHas("Authors", func(query contractsorm.Query) contractsorm.Query {
						return query.Where("name", "goravel")
					})
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_authors` WHERE `name` = ? AND `relation_authors`.`relation_book_id` = `relation_books`.`id`) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name:       "WhereHas with a nil callback",
			query:      func() contractsorm.Query { return users().WhereHas("Books", nil) },
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: nil,
		},
		{
			name:       "Has applies the related global scopes",
			query:      func() contractsorm.Query { return newQuery().Model(&relationScopedUser{}).Has("Books") },
			expectSql:  "SELECT * FROM `relation_scoped_users` WHERE EXISTS (SELECT 1 FROM `relation_scoped_books` WHERE `published` = ? AND `relation_scoped_books`.`relation_user_id` = `relation_scoped_users`.`id`)",
			expectVars: []any{true},
		},
		{
			name: "OrWhereHas",
			query: func() contractsorm.Query {
				return users().Where("name", "goravel").OrWhereHas("Books", published)
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? OR EXISTS (SELECT 1 FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel", true},
		},
		{
			name:       "WhereDoesntHave",
			query:      func() contractsorm.Query { return users().WhereDoesntHave("Books", published) },
			expectSql:  "SELECT * FROM `relation_users` WHERE NOT EXISTS (SELECT 1 FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{true},
		},
		{
			name: "OrWhereDoesntHave",
			query: func() contractsorm.Query {
				return users().Where("name", "goravel").OrWhereDoesntHave("Books", published)
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE `name` = ? OR NOT EXISTS (SELECT 1 FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel", true},
		},
		{
			name:      "Has with a self referencing has many",
			query:     func() contractsorm.Query { return newQuery().Model(&relationNode{}).Has("Children") },
			expectSql: "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `children_relation_nodes` WHERE `children_relation_nodes`.`parent_id` = `relation_nodes`.`id` AND `children_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
		},
		{
			name:      "Has with a self referencing belongs to",
			query:     func() contractsorm.Query { return newQuery().Model(&relationNode{}).Has("Parent") },
			expectSql: "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `parent_relation_nodes` WHERE `parent_relation_nodes`.`id` = `relation_nodes`.`parent_id` AND `parent_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
		},
		{
			name:      "Has with a self referencing many to many",
			query:     func() contractsorm.Query { return newQuery().Model(&relationNode{}).Has("Friends") },
			expectSql: "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `friends_relation_nodes` INNER JOIN `relation_node_friends` ON `relation_node_friends`.`friend_id` = `friends_relation_nodes`.`id` WHERE `relation_node_friends`.`relation_node_id` = `relation_nodes`.`id` AND `friends_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
		},
		{
			name:       "Has with a nested self referencing relation",
			query:      func() contractsorm.Query { return newQuery().Model(&relationNode{}).Has("Children.Children", ">=", 2) },
			expectSql:  "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `children_relation_nodes` WHERE (SELECT COUNT(*) FROM `relation_nodes` WHERE `relation_nodes`.`parent_id` = `children_relation_nodes`.`id` AND `relation_nodes`.`deleted_at` IS NULL) >= ? AND `children_relation_nodes`.`parent_id` = `relation_nodes`.`id` AND `children_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
			expectVars: []any{int64(2)},
		},
		{
			name: "WhereHas with a self referencing relation",
			query: func() contractsorm.Query {
				return newQuery().Model(&relationNode{}).WhereHas("Children", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("id > ?", 1)
				})
			},
			expectSql:  "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `children_relation_nodes` WHERE id > ? AND `children_relation_nodes`.`parent_id` = `relation_nodes`.`id` AND `children_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
			expectVars: []any{1},
		},
		{
			name: "Has inside a group of a WhereHas callback on a self referencing relation",
			query: func() contractsorm.Query {
				return newQuery().Model(&relationNode{}).WhereHas("Children", func(query contractsorm.Query) contractsorm.Query {
					return query.Where(func(query contractsorm.Query) contractsorm.Query {
						return query.Has("Parent")
					})
				})
			},
			expectSql: "SELECT * FROM `relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` AS `children_relation_nodes` WHERE EXISTS (SELECT 1 FROM `relation_nodes` WHERE `relation_nodes`.`id` = `children_relation_nodes`.`parent_id` AND `relation_nodes`.`deleted_at` IS NULL) AND `children_relation_nodes`.`parent_id` = `relation_nodes`.`id` AND `children_relation_nodes`.`deleted_at` IS NULL) AND `relation_nodes`.`deleted_at` IS NULL",
		},
		{
			name: "Has inside a group on a table alias",
			query: func() contractsorm.Query {
				return users().Table("relation_users AS u").Where(func(query contractsorm.Query) contractsorm.Query {
					return query.Has("Books")
				})
			},
			expectSql: "SELECT * FROM relation_users AS u WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `u`.`id`)",
		},
		{
			name:       "Has with a count on a self referencing relation",
			query:      func() contractsorm.Query { return newQuery().Model(&relationNode{}).Has("Children", ">=", 2) },
			expectSql:  "SELECT * FROM `relation_nodes` WHERE (SELECT COUNT(*) FROM `relation_nodes` AS `children_relation_nodes` WHERE `children_relation_nodes`.`parent_id` = `relation_nodes`.`id` AND `children_relation_nodes`.`deleted_at` IS NULL) >= ? AND `relation_nodes`.`deleted_at` IS NULL",
			expectVars: []any{int64(2)},
		},
		{
			name: "WhereHas groups an OR of the callback",
			query: func() contractsorm.Query {
				return users().WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("published", true).OrWhere("featured", true)
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_books` WHERE (`published` = ? OR `featured` = ?) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{true, true},
		},
		{
			name:       "Has groups an OR of a related global scope",
			query:      func() contractsorm.Query { return newQuery().Model(&relationOrScopedUser{}).Has("Books") },
			expectSql:  "SELECT * FROM `relation_or_scoped_users` WHERE EXISTS (SELECT 1 FROM `relation_or_scoped_books` WHERE (`published` = ? OR `featured` = ?) AND `relation_or_scoped_books`.`relation_user_id` = `relation_or_scoped_users`.`id`)",
			expectVars: []any{true, true},
		},
		{
			name:      "DoesntHave with a nested relation",
			query:     func() contractsorm.Query { return users().DoesntHave("Books.Authors") },
			expectSql: "SELECT * FROM `relation_users` WHERE NOT EXISTS (SELECT 1 FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_authors` WHERE `relation_authors`.`relation_book_id` = `relation_books`.`id`) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
		},
		{
			name: "WhereDoesntHave with a nested relation",
			query: func() contractsorm.Query {
				return users().WhereDoesntHave("Books.Authors", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "goravel")
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE NOT EXISTS (SELECT 1 FROM `relation_books` WHERE EXISTS (SELECT 1 FROM `relation_authors` WHERE `name` = ? AND `relation_authors`.`relation_book_id` = `relation_books`.`id`) AND `relation_books`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"goravel"},
		},
		{
			name: "WhereHas with many to many",
			query: func() contractsorm.Query {
				return users().WhereHas("Roles", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "admin")
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_roles` INNER JOIN `relation_role_users` ON `relation_role_users`.`relation_role_id` = `relation_roles`.`id` WHERE `name` = ? AND `relation_role_users`.`relation_user_id` = `relation_users`.`id`)",
			expectVars: []any{"admin"},
		},
		{
			name: "WhereHas with polymorphic has many",
			query: func() contractsorm.Query {
				return users().WhereHas("Phones", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("number", "123")
				})
			},
			expectSql:  "SELECT * FROM `relation_users` WHERE EXISTS (SELECT 1 FROM `relation_phones` WHERE `number` = ? AND `relation_phones`.`phoneable_type` = ? AND `relation_phones`.`phoneable_id` = `relation_users`.`id`)",
			expectVars: []any{"123", "relation_users"},
		},
		{
			name:       "Has with the != operator",
			query:      func() contractsorm.Query { return users().Has("Books", "!=", 2) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) != ?",
			expectVars: []any{int64(2)},
		},
		{
			name:       "Has with the <> operator",
			query:      func() contractsorm.Query { return users().Has("Books", "<>", 2) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) <> ?",
			expectVars: []any{int64(2)},
		},
		{
			name:       "Has with the <= operator",
			query:      func() contractsorm.Query { return users().Has("Books", "<=", 2) },
			expectSql:  "SELECT * FROM `relation_users` WHERE (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) <= ?",
			expectVars: []any{int64(2)},
		},
		{
			name:      "Has with a self referencing relation on a mixed case table",
			query:     func() contractsorm.Query { return newQuery().Model(&relationMixedNode{}).Has("Children") },
			expectSql: "SELECT * FROM `Nodes` WHERE EXISTS (SELECT 1 FROM `Nodes` AS `children_Nodes` WHERE `children_Nodes`.`parent_id` = `Nodes`.`id`)",
		},
		{
			name:        "Has with an unsigned count above the maximum",
			query:       func() contractsorm.Query { return users().Has("Books", "=", uint64(math.MaxInt64)+1) },
			expectError: errors.OrmQueryInvalidRelationCount,
		},
		{
			name:        "Has with a related model on another connection",
			query:       func() contractsorm.Query { return newQuery().Model(&relationConnectionUser{}).Has("Books") },
			expectError: errors.OrmQueryRelationDifferentConnection.Args("Books", "other", ""),
		},
		{
			name:        "Has with a negative count",
			query:       func() contractsorm.Query { return users().Has("Books", ">=", -1) },
			expectError: errors.OrmQueryInvalidRelationCount,
		},
		{
			name:        "Has with an empty relation",
			query:       func() contractsorm.Query { return users().Has("") },
			expectError: errors.OrmQueryEmptyRelation,
		},
		{
			name:        "Has with an unknown relation",
			query:       func() contractsorm.Query { return users().Has("Posts") },
			expectError: errors.OrmQueryRelationNotFound.Args("Posts", "relationUser"),
		},
		{
			name:        "Has with an unknown nested relation",
			query:       func() contractsorm.Query { return users().Has("Books.Posts") },
			expectError: errors.OrmQueryRelationNotFound.Args("Posts", "relationBook"),
		},
		{
			name:        "Has without a model",
			query:       func() contractsorm.Query { return newQuery().Table("users").Has("Books") },
			dest:        &[]map[string]any{},
			expectError: errors.OrmQueryRelationWithoutModel.Args("Books"),
		},
		{
			name:        "Has with only an operator",
			query:       func() contractsorm.Query { return users().Has("Books", ">") },
			expectError: errors.OrmQueryInvalidRelationCount,
		},
		{
			name:        "Has with an unknown operator",
			query:       func() contractsorm.Query { return users().Has("Books", "like", 1) },
			expectError: errors.OrmQueryInvalidRelationCount,
		},
		{
			name:        "Has with a count that is not an integer",
			query:       func() contractsorm.Query { return users().Has("Books", ">", "1") },
			expectError: errors.OrmQueryInvalidRelationCount,
		},
		{
			name: "WhereHas with an invalid callback",
			query: func() contractsorm.Query {
				return users().WhereHas("Books", func(contractsorm.Query) contractsorm.Query { return nil })
			},
			expectError: errors.OrmQueryInvalidSubQuery,
		},
		{
			name: "WhereHas propagates the callback error",
			query: func() contractsorm.Query {
				return users().WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
					return query.WhereColumn("title")
				})
			},
			expectError: errors.DatabaseInvalidArgumentNumber.Args(0, "1 or 2"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dest := test.dest
			if dest == nil {
				dest = &[]map[string]any{}
			}

			query := test.query().(*Query)
			if test.dest != nil {
				query = query.dest(dest)
			}
			query = query.buildConditions()
			if test.expectError != nil {
				assert.EqualError(t, query.instance.Error, test.expectError.Error())
				return
			}

			statement := query.instance.Find(dest).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectSql, statement.SQL.String())
			if len(test.expectVars) == 0 {
				assert.Empty(t, statement.Vars)
			} else {
				assert.Equal(t, test.expectVars, statement.Vars)
			}
		})
	}
}

func TestRelationExistenceWithTablePrefix(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{
		DryRun:         true,
		NamingStrategy: schema.NamingStrategy{TablePrefix: "pre_"},
	})
	require.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{Prefix: "pre_"}, db, nil, nil, nil, nil, nil)
	}

	tests := []struct {
		name      string
		query     contractsorm.Query
		expectSql string
	}{
		{
			name:      "a self referencing relation",
			query:     newQuery().Model(&relationNode{}).Has("Children"),
			expectSql: "SELECT * FROM `pre_relation_nodes` WHERE EXISTS (SELECT 1 FROM `pre_relation_nodes` AS `children_pre_relation_nodes` WHERE `children_pre_relation_nodes`.`parent_id` = `pre_relation_nodes`.`id` AND `children_pre_relation_nodes`.`deleted_at` IS NULL) AND `pre_relation_nodes`.`deleted_at` IS NULL",
		},
		{
			name:      "a relation on another table",
			query:     newQuery().Model(&relationUser{}).Has("Books"),
			expectSql: "SELECT * FROM `pre_relation_users` WHERE EXISTS (SELECT 1 FROM `pre_relation_books` WHERE `pre_relation_books`.`relation_user_id` = `pre_relation_users`.`id`)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dest := &[]map[string]any{}
			query := test.query.(*Query).buildConditions()
			require.NoError(t, query.instance.Error)

			statement := query.instance.Find(dest).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectSql, statement.SQL.String())
		})
	}
}
