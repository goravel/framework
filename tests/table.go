package tests

import (
	"github.com/goravel/framework/contracts/database/driver"
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
	TestTableJsonData
	TestTableGlobalScopes
	TestTableUuidEntities
	TestTableUlidEntities
	TestTableMorphableEntities
	TestTableUuidMorphableEntities
	TestTableUlidMorphableEntities
)

type testTables struct {
	driver  string
	grammar driver.Grammar
}

func newTestTables(driver string, grammar driver.Grammar) *testTables {
	return &testTables{driver: driver, grammar: grammar}
}

func (r *testTables) All() map[TestTable]func() ([]string, error) {
	return map[TestTable]func() ([]string, error){
		TestTableAddresses:             r.addresses,
		TestTableAuthors:               r.authors,
		TestTableBooks:                 r.books,
		TestTableHouses:                r.houses,
		TestTablePeoples:               r.peoples,
		TestTablePhones:                r.phones,
		TestTableProducts:              r.products,
		TestTableReviews:               r.reviews,
		TestTableRoles:                 r.roles,
		TestTableRoleUser:              r.roleUser,
		TestTableUsers:                 r.users,
		TestTableUser:                  r.user,
		TestTableSchema:                r.schemas,
		TestTableJsonData:              r.jsonData,
		TestTableGlobalScopes:          r.globalScopes,
		TestTableUuidEntities:          r.uuidEntities,
		TestTableUlidEntities:          r.ulidEntities,
		TestTableMorphableEntities:     r.morphableEntities,
		TestTableUuidMorphableEntities: r.uuidMorphableEntities,
		TestTableUlidMorphableEntities: r.ulidMorphableEntities,
	}
}

func (r *testTables) peoples() ([]string, error) {
	dropSql, err := r.dropSql("peoples")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "peoples")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("body")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) reviews() ([]string, error) {
	dropSql, err := r.dropSql("reviews")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "reviews")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("body")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) products() ([]string, error) {
	dropSql, err := r.dropSql("products")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "products")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.Integer("weight").Nullable()
	blueprint.Integer("height").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) users() ([]string, error) {
	dropSql, err := r.dropSql("users")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "users")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("bio").Nullable()
	blueprint.String("avatar").Nullable()
	blueprint.Float("ratio").Nullable().Default(0)
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) user() ([]string, error) {
	dropSql, err := r.dropSql("user")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "user")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("bio").Nullable()
	blueprint.String("avatar").Nullable()
	blueprint.Float("ratio").Nullable().Default(0)
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) addresses() ([]string, error) {
	dropSql, err := r.dropSql("addresses")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "addresses")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("user_id").Nullable()
	blueprint.String("name")
	blueprint.String("province").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) books() ([]string, error) {
	dropSql, err := r.dropSql("books")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "books")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("user_id").Nullable()
	blueprint.String("name")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) authors() ([]string, error) {
	dropSql, err := r.dropSql("authors")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "authors")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("book_id").Nullable()
	blueprint.String("name")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) roles() ([]string, error) {
	dropSql, err := r.dropSql("roles")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "roles")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.String("avatar").Nullable()
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) houses() ([]string, error) {
	dropSql, err := r.dropSql("houses")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "houses")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UnsignedBigInteger("houseable_id")
	blueprint.String("houseable_type")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) phones() ([]string, error) {
	dropSql, err := r.dropSql("phones")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "phones")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UnsignedBigInteger("phoneable_id")
	blueprint.String("phoneable_type")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) roleUser() ([]string, error) {
	dropSql, err := r.dropSql("role_user")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "role_user")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.UnsignedBigInteger("role_id")
	blueprint.UnsignedBigInteger("user_id")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) schemas() ([]string, error) {
	dropSql, err := r.dropSql("goravel.schemas")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "goravel.schemas")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) jsonData() ([]string, error) {
	dropSql, err := r.dropSql("json_data")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "json_data")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.Json("data")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) globalScopes() ([]string, error) {
	dropSql, err := r.dropSql("global_scopes")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "global_scopes")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.Timestamps()
	blueprint.SoftDeletes()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) uuidEntities() ([]string, error) {
	dropSql, err := r.dropSql("uuid_entities")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "uuid_entities")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.Uuid("uuid")
	blueprint.String("name")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) ulidEntities() ([]string, error) {
	dropSql, err := r.dropSql("ulid_entities")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "ulid_entities")
	blueprint.Create()
	blueprint.Ulid("id")
	blueprint.Primary("id")
	blueprint.String("name")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) morphableEntities() ([]string, error) {
	dropSql, err := r.dropSql("morphable_entities")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "morphable_entities")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.NumericMorphs("morphable")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) uuidMorphableEntities() ([]string, error) {
	dropSql, err := r.dropSql("uuid_morphable_entities")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "uuid_morphable_entities")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UuidMorphs("morphable")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) ulidMorphableEntities() ([]string, error) {
	dropSql, err := r.dropSql("ulid_morphable_entities")
	if err != nil {
		return nil, err
	}

	blueprint := schema.NewBlueprint(nil, "", "ulid_morphable_entities")
	blueprint.Create()
	blueprint.BigIncrements("id")
	blueprint.String("name")
	blueprint.UlidMorphs("morphable")
	blueprint.Timestamps()

	createSql, err := blueprint.ToSql(r.grammar)
	if err != nil {
		return nil, err
	}

	return append(dropSql, createSql...), nil
}

func (r *testTables) dropSql(table string) ([]string, error) {
	blueprint := schema.NewBlueprint(nil, "", table)
	blueprint.DropIfExists()

	return blueprint.ToSql(r.grammar)
}
