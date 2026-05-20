package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// newRelQuery builds a Query backed by a stub gorm.DB. The DB is non-functional but is enough for
// the queueing methods (Has / WhereHas / WithCount / With / ...) which only mutate the
// Conditions value and never execute SQL.
func newRelQuery(t *testing.T) *Query {
	t.Helper()
	db := newStubGormDB(t)
	conditions := Conditions{}
	return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, &conditions)
}

func toQuery(q contractsorm.Query) *Query {
	return q.(*Query)
}

// --- Existence queueing ----------------------------------------------------

func TestHasQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.Has("Books"))
	assert.Len(t, got.conditions.relations, 1)
	rel := got.conditions.relations[0]
	assert.Equal(t, "Books", rel.relation)
	assert.Equal(t, ">=", rel.operator)
	assert.Equal(t, 1, rel.count)
	assert.Equal(t, "and", rel.conjunction)
	assert.Nil(t, rel.callback)
}

func TestHasWithCallbackOpAndCount(t *testing.T) {
	q := newRelQuery(t)
	cb := contractsorm.RelationCallback(func(query contractsorm.Query) contractsorm.Query { return query })
	got := toQuery(q.Has("Books", cb, ">", 3))
	assert.Len(t, got.conditions.relations, 1)
	rel := got.conditions.relations[0]
	assert.Equal(t, ">", rel.operator)
	assert.Equal(t, 3, rel.count)
	assert.NotNil(t, rel.callback)
}

func TestOrHasQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.OrHas("Books"))
	assert.Equal(t, "or", got.conditions.relations[0].conjunction)
}

func TestDoesntHaveQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.DoesntHave("Books"))
	rel := got.conditions.relations[0]
	assert.Equal(t, "<", rel.operator)
	assert.Equal(t, 1, rel.count)
	assert.Equal(t, "and", rel.conjunction)
}

func TestOrDoesntHaveQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.OrDoesntHave("Books"))
	assert.Equal(t, "or", got.conditions.relations[0].conjunction)
	assert.Equal(t, "<", got.conditions.relations[0].operator)
}

func TestWhereHasFamilyQueueing(t *testing.T) {
	cb := contractsorm.RelationCallback(func(q contractsorm.Query) contractsorm.Query { return q })
	tests := []struct {
		name        string
		invoke      func(*Query) contractsorm.Query
		conjunction string
		operator    string
	}{
		{"WhereHas", func(q *Query) contractsorm.Query { return q.WhereHas("Books", cb) }, "and", ">="},
		{"OrWhereHas", func(q *Query) contractsorm.Query { return q.OrWhereHas("Books", cb) }, "or", ">="},
		{"WhereDoesntHave", func(q *Query) contractsorm.Query { return q.WhereDoesntHave("Books", cb) }, "and", "<"},
		{"OrWhereDoesntHave", func(q *Query) contractsorm.Query { return q.OrWhereDoesntHave("Books", cb) }, "or", "<"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := newRelQuery(t)
			got := toQuery(tc.invoke(q))
			assert.Len(t, got.conditions.relations, 1)
			assert.Equal(t, tc.conjunction, got.conditions.relations[0].conjunction)
			assert.Equal(t, tc.operator, got.conditions.relations[0].operator)
			assert.NotNil(t, got.conditions.relations[0].callback)
		})
	}
}

func TestHasInvalidArgErrorPropagated(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.Has("Books", 1.23))
	// invalid arg path returns a fresh query carrying an error on its instance.
	assert.Error(t, got.instance.Error)
}

// --- Morph queueing --------------------------------------------------------

func TestHasMorphQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.HasMorph("Houseable", []any{&relUser{}}))
	assert.Len(t, got.conditions.relations, 1)
	rel := got.conditions.relations[0]
	assert.Equal(t, "Houseable", rel.relation)
	assert.Len(t, rel.morphTypes, 1)
}

func TestMorphFamilyQueueing(t *testing.T) {
	mcb := contractsorm.MorphRelationCallback(func(q contractsorm.Query, _ string) contractsorm.Query { return q })
	tests := []struct {
		name        string
		invoke      func(*Query) contractsorm.Query
		conjunction string
		operator    string
	}{
		{"OrHasMorph", func(q *Query) contractsorm.Query { return q.OrHasMorph("X", []any{&relUser{}}) }, "or", ">="},
		{"DoesntHaveMorph", func(q *Query) contractsorm.Query { return q.DoesntHaveMorph("X", []any{&relUser{}}) }, "and", "<"},
		{"OrDoesntHaveMorph", func(q *Query) contractsorm.Query { return q.OrDoesntHaveMorph("X", []any{&relUser{}}) }, "or", "<"},
		{"WhereHasMorph", func(q *Query) contractsorm.Query { return q.WhereHasMorph("X", []any{&relUser{}}, mcb) }, "and", ">="},
		{"OrWhereHasMorph", func(q *Query) contractsorm.Query { return q.OrWhereHasMorph("X", []any{&relUser{}}, mcb) }, "or", ">="},
		{"WhereDoesntHaveMorph", func(q *Query) contractsorm.Query { return q.WhereDoesntHaveMorph("X", []any{&relUser{}}) }, "and", "<"},
		{"OrWhereDoesntHaveMorph", func(q *Query) contractsorm.Query { return q.OrWhereDoesntHaveMorph("X", []any{&relUser{}}) }, "or", "<"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := newRelQuery(t)
			got := toQuery(tc.invoke(q))
			assert.Len(t, got.conditions.relations, 1)
			assert.Equal(t, tc.conjunction, got.conditions.relations[0].conjunction)
			assert.Equal(t, tc.operator, got.conditions.relations[0].operator)
		})
	}
}

func TestHasMorphEmptyTypesError(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.HasMorph("X", nil))
	assert.True(t, errors.Is(got.instance.Error, errors.OrmRelationMorphTypesEmpty))
}

func TestHasMorphInvalidArgError(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.HasMorph("X", []any{&relUser{}}, 3.14))
	assert.Error(t, got.instance.Error)
}

// --- Aggregate / WithCount / WithExists queueing --------------------------

func TestWithAggregate(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithAggregate("Books", "id", "max"))
	assert.Len(t, got.conditions.selectSubs, 1)
	sub := got.conditions.selectSubs[0]
	assert.Equal(t, "Books", sub.relation)
	assert.Equal(t, "id", sub.column)
	assert.Equal(t, "max", sub.function)
	assert.Equal(t, "books_max_id", sub.alias)
}

func TestWithAggregateInvalidFn(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithAggregate("Books", "*", "median"))
	assert.True(t, errors.Is(got.instance.Error, errors.OrmRelationInvalidAggregate))
}

func TestWithAggregateInvalidArg(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithAggregate("Books", "*", "count", 3.14))
	assert.True(t, errors.Is(got.instance.Error, errors.OrmRelationInvalidArgument))
}

func TestWithCountStringAndStruct(t *testing.T) {
	q := newRelQuery(t)
	cb := contractsorm.RelationCallback(func(q contractsorm.Query) contractsorm.Query { return q })
	got := toQuery(q.WithCount("Books", contractsorm.RelationCount{Name: "Roles", Alias: "rcount", Callback: cb}))
	assert.Len(t, got.conditions.selectSubs, 2)
	assert.Equal(t, "books_count", got.conditions.selectSubs[0].alias)
	assert.Equal(t, "rcount", got.conditions.selectSubs[1].alias)
	assert.NotNil(t, got.conditions.selectSubs[1].callback)
}

func TestWithCountInvalidArg(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithCount(123))
	assert.True(t, errors.Is(got.instance.Error, errors.OrmRelationInvalidArgument))
}

func TestWithMaxMinSumAvg(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithMax("Books", "id"))
	got = toQuery(got.WithMin("Books", "id"))
	got = toQuery(got.WithSum("Books", "id"))
	got = toQuery(got.WithAvg("Books", "id"))
	assert.Len(t, got.conditions.selectSubs, 4)
	assert.Equal(t, "max", got.conditions.selectSubs[0].function)
	assert.Equal(t, "min", got.conditions.selectSubs[1].function)
	assert.Equal(t, "sum", got.conditions.selectSubs[2].function)
	assert.Equal(t, "avg", got.conditions.selectSubs[3].function)
}

func TestWithExists(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.WithExists("Books", "Roles"))
	assert.Len(t, got.conditions.selectSubs, 2)
	assert.Equal(t, "exists", got.conditions.selectSubs[0].function)
	assert.Equal(t, "exists", got.conditions.selectSubs[1].function)
	assert.Equal(t, "books_exists", got.conditions.selectSubs[0].alias)
}

// --- Eager-load queueing ---------------------------------------------------

func TestWithQueueing(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.With("Books", "Roles"))
	assert.Len(t, got.conditions.eagerLoad, 2)
	assert.Equal(t, "Books", got.conditions.eagerLoad[0].relation)
	assert.Equal(t, "Roles", got.conditions.eagerLoad[1].relation)
}

func TestWithInvalidArg(t *testing.T) {
	q := newRelQuery(t)
	got := toQuery(q.With(3.14))
	assert.True(t, errors.Is(got.instance.Error, errors.OrmEagerLoadInvalidArgument))
}

func TestWithout(t *testing.T) {
	q := newRelQuery(t)
	q1 := toQuery(q.With("Books", "Roles", "Logo"))
	q2 := toQuery(q1.Without("Roles"))
	assert.Len(t, q2.conditions.eagerLoad, 2)
	for _, e := range q2.conditions.eagerLoad {
		assert.NotEqual(t, "Roles", e.relation)
	}
}

func TestWithoutNoOps(t *testing.T) {
	q := newRelQuery(t)
	// no eager loads queued -> returns same query
	q1 := q.Without("Books")
	assert.Same(t, q, q1.(*Query))

	// no relation names -> returns same query
	q2 := toQuery(q.With("Books"))
	q3 := q2.Without()
	assert.Same(t, q2, q3.(*Query))
}

func TestWithOnly(t *testing.T) {
	q := newRelQuery(t)
	q1 := toQuery(q.With("Books", "Roles"))
	q2 := toQuery(q1.WithOnly("Logo"))
	assert.Len(t, q2.conditions.eagerLoad, 1)
	assert.Equal(t, "Logo", q2.conditions.eagerLoad[0].relation)
}

// --- Pure helpers ---------------------------------------------------------

func TestParseRelationArgsAllShapes(t *testing.T) {
	cb := contractsorm.RelationCallback(func(q contractsorm.Query) contractsorm.Query { return q })
	cases := []struct {
		name  string
		args  []any
		op    string
		count int
		hasCb bool
	}{
		{"empty", nil, ">=", 1, false},
		{"callback", []any{cb}, ">=", 1, true},
		{"callback as func", []any{func(q contractsorm.Query) contractsorm.Query { return q }}, ">=", 1, true},
		{"op only", []any{">"}, ">", 1, false},
		{"op + count", []any{">", 5}, ">", 5, false},
		{"int64 count", []any{int64(2)}, ">=", 2, false},
		{"nil arg ignored", []any{nil, "<"}, "<", 1, false},
		{"callback + op + count", []any{cb, ">=", 3}, ">=", 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotCb, op, count, err := parseRelationArgs(tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.op, op)
			assert.Equal(t, tc.count, count)
			assert.Equal(t, tc.hasCb, gotCb != nil)
		})
	}
}

func TestParseRelationArgsInvalid(t *testing.T) {
	_, _, _, err := parseRelationArgs([]any{3.14})
	assert.True(t, errors.Is(err, errors.OrmRelationInvalidArgument))
}

func TestParseMorphRelationArgs(t *testing.T) {
	cb := contractsorm.RelationCallback(func(q contractsorm.Query) contractsorm.Query { return q })
	mcb := contractsorm.MorphRelationCallback(func(q contractsorm.Query, _ string) contractsorm.Query { return q })

	gotCb, gotMcb, op, count, err := parseMorphRelationArgs([]any{cb, ">", 3})
	assert.NoError(t, err)
	assert.NotNil(t, gotCb)
	assert.Nil(t, gotMcb)
	assert.Equal(t, ">", op)
	assert.Equal(t, 3, count)

	gotCb, gotMcb, _, _, err = parseMorphRelationArgs([]any{mcb})
	assert.NoError(t, err)
	assert.Nil(t, gotCb)
	assert.NotNil(t, gotMcb)

	gotCb, gotMcb, _, _, err = parseMorphRelationArgs([]any{func(q contractsorm.Query, _ string) contractsorm.Query { return q }})
	assert.NoError(t, err)
	assert.Nil(t, gotCb)
	assert.NotNil(t, gotMcb)

	gotCb, _, _, _, err = parseMorphRelationArgs([]any{func(q contractsorm.Query) contractsorm.Query { return q }})
	assert.NoError(t, err)
	assert.NotNil(t, gotCb)

	_, _, _, _, err = parseMorphRelationArgs([]any{nil, int64(7), "="})
	assert.NoError(t, err)

	_, _, _, _, err = parseMorphRelationArgs([]any{3.14})
	assert.True(t, errors.Is(err, errors.OrmRelationInvalidArgument))
}

func TestShouldUseExists(t *testing.T) {
	cases := []struct {
		op    string
		count int
		want  bool
	}{
		{">=", 1, true},
		{"<", 1, true},
		{">=", 2, false},
		{"<", 2, false},
		{"=", 1, false},
		{">", 1, false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, shouldUseExists(tc.op, tc.count), "op=%s count=%d", tc.op, tc.count)
	}
}

func TestValidAggregateFn(t *testing.T) {
	for _, fn := range []string{"count", "max", "min", "sum", "avg", "exists"} {
		assert.True(t, validAggregateFn(fn), fn)
	}
	for _, fn := range []string{"", "median", "stddev"} {
		assert.False(t, validAggregateFn(fn), fn)
	}
}

func TestAggregateAlias(t *testing.T) {
	cases := []struct {
		relation string
		fn       string
		col      string
		want     string
	}{
		{"Books", "count", "*", "books_count"},
		{"Books", "count", "", "books_count"},
		{"Books.Author", "count", "*", "books_author_count"},
		{"BooksAuthor", "max", "id", "books_author_max_id"},
		{"books_author", "sum", "price", "books_author_sum_price"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, aggregateAlias(tc.relation, tc.fn, tc.col))
	}
}

func TestQuoteIdent(t *testing.T) {
	assert.Equal(t, "", quoteIdent(""))
	assert.Equal(t, "users", quoteIdent("users"))
	// expressions left alone (contain space, parens, dot)
	assert.Equal(t, "users.id", quoteIdent("users.id"))
	assert.Equal(t, "COUNT(*)", quoteIdent("COUNT(*)"))
}

func TestParentTable(t *testing.T) {
	q := newRelQuery(t)
	tbl := q.parentTable(&relUser{})
	assert.Equal(t, "rel_users", tbl)
	// invalid model returns ""
	assert.Equal(t, "", q.parentTable("not-a-model"))
}

func TestParentModelFromConditions(t *testing.T) {
	q := newRelQuery(t)
	q.conditions.model = &relUser{}
	assert.NotNil(t, q.parentModel())

	q2 := newRelQuery(t)
	q2.conditions.dest = &[]relUser{}
	assert.NotNil(t, q2.parentModel())

	q3 := newRelQuery(t)
	assert.Nil(t, q3.parentModel())
}

func TestFreshSession(t *testing.T) {
	q := newRelQuery(t)
	s := q.freshSession()
	assert.NotNil(t, s)
}

func TestWrapReturnsQuery(t *testing.T) {
	q := newRelQuery(t)
	wrapped := q.wrap(q.instance)
	assert.NotNil(t, wrapped)
	assert.NotNil(t, wrapped.instance)
}
