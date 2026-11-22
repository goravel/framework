package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormio "gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

func TestWithCount(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	require.NoError(t, err)

	users := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(&relationUser{})
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
			name:      "WithCount",
			query:     func() contractsorm.Query { return users().WithCount("Books") },
			expectSql: "SELECT `relation_users`.*, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count` FROM `relation_users`",
		},
		{
			name:       "WithCount with an alias and a callback",
			query:      func() contractsorm.Query { return users().WithCount("Books as published_count", published) },
			expectSql:  "SELECT `relation_users`.*, (SELECT COUNT(*) FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `published_count` FROM `relation_users`",
			expectVars: []any{true},
		},
		{
			name: "WithCount several times",
			query: func() contractsorm.Query {
				return users().WithCount("Books").WithCount("Books AS published_count", published)
			},
			expectSql:  "SELECT `relation_users`.*, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count`, (SELECT COUNT(*) FROM `relation_books` WHERE `published` = ? AND `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `published_count` FROM `relation_users`",
			expectVars: []any{true},
		},
		{
			name:      "WithCount keeps the selected columns",
			query:     func() contractsorm.Query { return users().Select("id", "name").WithCount("Books") },
			expectSql: "SELECT id, name, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count` FROM `relation_users`",
		},
		{
			name:       "WithCount keeps the raw select",
			query:      func() contractsorm.Query { return users().SelectRaw("id, ? AS kind", "user").WithCount("Books") },
			expectSql:  "SELECT id, ? AS kind, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count` FROM `relation_users`",
			expectVars: []any{"user"},
		},
		{
			name:       "WithCount with conditions on the parent",
			query:      func() contractsorm.Query { return users().Where("name", "goravel").WithCount("Books") },
			expectSql:  "SELECT `relation_users`.*, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count` FROM `relation_users` WHERE `name` = ?",
			expectVars: []any{"goravel"},
		},
		{
			name:      "WithCount reads into the destination struct",
			query:     func() contractsorm.Query { return users().WithCount("Books") },
			dest:      &[]relationUser{},
			expectSql: "SELECT `relation_users`.*, (SELECT COUNT(*) FROM `relation_books` WHERE `relation_books`.`relation_user_id` = `relation_users`.`id`) AS `books_count` FROM `relation_users`",
		},
		{
			name:        "WithCount without a field for the column",
			query:       func() contractsorm.Query { return users().WithCount("Roles") },
			dest:        &[]relationUser{},
			expectError: errors.OrmQueryRelationCountNoField.Args("relationUser", "roles_count"),
		},
		{
			name:        "WithCount with an empty relation",
			query:       func() contractsorm.Query { return users().WithCount("") },
			expectError: errors.OrmQueryEmptyRelation,
		},
		{
			name:        "WithCount with an unknown relation",
			query:       func() contractsorm.Query { return users().WithCount("Posts as books_count") },
			expectError: errors.OrmQueryRelationNotFound.Args("Posts", "relationUser"),
		},
		{
			name: "WithCount with an invalid callback",
			query: func() contractsorm.Query {
				return users().WithCount("Books", func(contractsorm.Query) contractsorm.Query { return nil })
			},
			expectError: errors.OrmQueryInvalidSubQuery,
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

func TestWithCountIsLeftOutOfCount(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	require.NoError(t, err)

	query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(&relationUser{}).WithCount("Books").(*Query)
	statement := buildSelectForCount(query).instance.Count(new(int64)).Statement

	assert.NoError(t, statement.Error)
	assert.Equal(t, "SELECT count(*) FROM `relation_users`", statement.SQL.String())
}
