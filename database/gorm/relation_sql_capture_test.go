package gorm

import (
	"fmt"
	"os"
	"testing"

	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// TestCapture_RelationSQL is a one-shot helper that writes the actual SQL generated for every
// relation type to /tmp/relation_sql.txt. Run with: go test -run TestCapture_RelationSQL.
// Then read /tmp/relation_sql.txt and paste the values into relation_sql_test.go.
func TestCapture_RelationSQL(t *testing.T) {
	if os.Getenv("CAPTURE_SQL") != "1" {
		t.Skip("set CAPTURE_SQL=1 to run")
	}

	f, err := os.Create("/tmp/relation_sql.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Logf("close: %v", err)
		}
	}()

	cases := []struct {
		name  string
		build func(t *testing.T) (contractsorm.Query, any)
	}{
		{"HasOne", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relUser{})
			return q.Related(&relUser{ID: 7}, "Profile"), &relProfile{}
		}},
		{"HasMany", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relUser{})
			return q.Related(&relUser{ID: 7}, "Books"), &[]relBook{}
		}},
		{"BelongsTo", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relBook{})
			return q.Related(&relBook{AuthorID: 5}, "Author"), &relUser{}
		}},
		{"Many2Many", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relUser{})
			return q.Related(&relUser{ID: 7}, "Roles"), &[]relRole{}
		}},
		{"MorphOne", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relUser{})
			return q.Related(&relUser{ID: 9}, "Logo"), &relLogo{}
		}},
		{"MorphMany", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relUser{})
			return q.Related(&relUser{ID: 9}, "Houses"), &[]relHouse{}
		}},
		{"MorphToMany", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &morphPost{})
			return q.Related(&morphPost{ID: 3}, "Tags"), &[]morphTag{}
		}},
		{"MorphedByMany", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &morphTag{})
			return q.Related(&morphTag{ID: 1}, "Posts"), &[]morphPost{}
		}},
		{"HasManyThrough", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relCountry{})
			return q.Related(&relCountry{ID: 1}, "Posts"), &[]relPost{}
		}},
		{"HasOneThrough", func(t *testing.T) (contractsorm.Query, any) {
			q := newRelQueryWith(t, &relCountry{})
			return q.Related(&relCountry{ID: 1}, "FirstPost"), &relPost{}
		}},
	}

	if _, err := fmt.Fprintln(f, "=== Related SQL ==="); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		q, dest := c.build(t)
		gq := q.(*Query)
		stmt := gq.buildConditions().instance.Session(&gormio.Session{DryRun: true}).Find(dest)
		if _, err := fmt.Fprintf(f, "[%s]\n%s\n\n", c.name, stmt.Statement.SQL.String()); err != nil {
			t.Fatal(err)
		}
	}

	existence := []struct {
		name     string
		model    any
		relation string
		dest     any
	}{
		{"HasOne", &relUser{}, "Profile", &relProfile{}},
		{"HasMany", &relUser{}, "Books", &[]relBook{}},
		{"BelongsTo", &relBook{}, "Author", &[]relUser{}},
		{"Many2Many", &relUser{}, "Roles", &[]relRole{}},
		{"MorphOne", &relUser{}, "Logo", &relLogo{}},
		{"MorphMany", &relUser{}, "Houses", &[]relHouse{}},
		{"MorphToMany", &morphPost{}, "Tags", &[]morphTag{}},
		{"MorphedByMany", &morphTag{}, "Posts", &[]morphPost{}},
		{"HasManyThrough", &relCountry{}, "Posts", &[]relPost{}},
		{"HasOneThrough", &relCountry{}, "FirstPost", &relPost{}},
	}

	if _, err := fmt.Fprintln(f, "=== ExistenceSubquery SQL ==="); err != nil {
		t.Fatal(err)
	}
	for _, c := range existence {
		q := newRelQueryWith(t, c.model)
		desc, err := resolveRelation(q.instance, c.model, c.relation)
		if err != nil {
			if _, err := fmt.Fprintf(f, "[%s] ERROR: %v\n\n", c.name, err); err != nil {
				t.Fatal(err)
			}
			continue
		}
		inner := q.compileExistenceSubquery(desc, nil)
		stmt := inner.Session(&gormio.Session{DryRun: true}).Find(c.dest)
		if _, err := fmt.Fprintf(f, "[%s]\n%s\n\n", c.name, stmt.Statement.SQL.String()); err != nil {
			t.Fatal(err)
		}
	}

	aggregates := []struct {
		name string
		sub  selectSub
	}{
		{"Count", selectSub{relation: "Books", column: "*", function: "count"}},
		{"Sum", selectSub{relation: "Books", column: "id", function: "sum"}},
		{"Max", selectSub{relation: "Books", column: "id", function: "max"}},
		{"Min", selectSub{relation: "Books", column: "id", function: "min"}},
		{"Avg", selectSub{relation: "Books", column: "id", function: "avg"}},
		{"Exists", selectSub{relation: "Books", column: "*", function: "exists"}},
	}

	if _, err := fmt.Fprintln(f, "=== AggregateSubquery SQL ==="); err != nil {
		t.Fatal(err)
	}
	for _, c := range aggregates {
		q := newRelQueryWith(t, &relUser{})
		desc, err := resolveRelation(q.instance, &relUser{}, "Books")
		if err != nil {
			if _, err := fmt.Fprintf(f, "[%s] ERROR: %v\n\n", c.name, err); err != nil {
				t.Fatal(err)
			}
			continue
		}
		inner := q.compileAggregateSubquery(desc, c.sub)
		stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
		if _, err := fmt.Fprintf(f, "[%s]\n%s\n\n", c.name, stmt.Statement.SQL.String()); err != nil {
			t.Fatal(err)
		}
	}

	t.Logf("Captured SQL written to /tmp/relation_sql.txt")
}
