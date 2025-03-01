package tests

import (
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema"
)

type TestTable int

const (
	TestTableAddresses TestTable = iota
	TestTableAuthors
	TestTableBooks
	TestTableHouses
	TestTablePeoples
	TestTablePhones
	TestTableProducts
	TestTableReviews
	TestTableRoles
	TestTableRoleUser
	TestTableUsers
	TestTableUser
	TestTableSchema
)

type testTables struct {
	driver  string
	grammar contractsschema.Grammar
}

func newTestTables(driver string, grammar contractsschema.Grammar) *testTables {
	return &testTables{driver: driver, grammar: grammar}
}

func (r *testTables) All() map[TestTable]func() ([]string, error) {
	return map[TestTable]func() ([]string, error){
		TestTableAddresses: r.addresses,
		TestTableAuthors:   r.authors,
		TestTableBooks:     r.books,
		TestTableHouses:    r.houses,
		TestTablePeoples:   r.peoples,
		TestTablePhones:    r.phones,
		TestTableProducts:  r.products,
		TestTableReviews:   r.reviews,
		TestTableRoles:     r.roles,
		TestTableRoleUser:  r.roleUser,
		TestTableUsers:     r.users,
		TestTableUser:      r.user,
		TestTableSchema:    r.schema,
	}
}

func (r *testTables) peoples() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "peoples")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("body")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) reviews() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "reviews")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("body")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) products() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "products")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.Integer("weight").Nullable()
	blueprint.Integer("height").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) users() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "users")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("bio").Nullable()
	blueprint.String("avatar").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) user() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "user")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("bio").Nullable()
	blueprint.String("avatar").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) addresses() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "addresses")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("user_id").Nullable()
	blueprint.String("name")
	blueprint.String("province").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) books() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "books")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("user_id").Nullable()
	blueprint.String("name")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) authors() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "authors")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("book_id").Nullable()
	blueprint.String("name")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) roles() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "roles")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("avatar").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) houses() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "houses")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UnsignedBigInteger("houseable_id")
	blueprint.String("houseable_type")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) phones() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "phones")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UnsignedBigInteger("phoneable_id")
	blueprint.String("phoneable_type")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) roleUser() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "role_user")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("role_id")
	blueprint.UnsignedBigInteger("user_id")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	return blueprint.ToSql(r.grammar)
}

func (r *testTables) schema() ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", "goravel.schemas")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.Timestamps()

	return blueprint.ToSql(r.grammar)
}
