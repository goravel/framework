package orm

import (
	"context"
	"database/sql"
)

type Orm interface {
	// Connection gets an Orm instance from the connection pool.
	Connection(name string) Orm
	// DB gets the underlying database connection.
	DB() (*sql.DB, error)
	// Query gets a new query builder instance.
	Query() Query
	// Factory gets a new factory instance for the given model name.
	Factory() Factory
	// Observe registers an observer with the Orm.
	Observe(model any, observer Observer)
	// Transaction runs a callback wrapped in a database transaction.
	Transaction(txFunc func(tx Transaction) error) error
	// WithContext sets the context to be used by the Orm.
	WithContext(ctx context.Context) Orm
}

type Transaction interface {
	Query
	// Commit commits the changes in a transaction.
	Commit() error
	// Rollback rolls back the changes in a transaction.
	Rollback() error
}

type Query interface {
	// Association gets an association instance by name.
	Association(association string) Association
	// Begin begins a new transaction
	Begin() (Transaction, error)
	// Driver gets the driver for the query.
	Driver() Driver
	// Count retrieve the "count" result of the query.
	Count(count *int64) error
	// Create inserts new record into the database.
	Create(value any) error
	// Cursor returns a cursor, use scan to iterate over the returned rows.
	Cursor() (chan Cursor, error)
	// Delete deletes records matching given conditions, if the conditions are empty will delete all records.
	Delete(value any, conds ...any) (*Result, error)
	// Distinct specifies distinct fields to query.
	Distinct(args ...any) Query
	// Exec executes raw sql
	Exec(sql string, values ...any) (*Result, error)
	// Exists returns true if matching records exist; otherwise, it returns false.
	Exists(exists *bool) error
	// Find finds records that match given conditions.
	Find(dest any, conds ...any) error
	// FindOrFail finds records that match given conditions or throws an error.
	FindOrFail(dest any, conds ...any) error
	// First finds record that match given conditions.
	First(dest any) error
	// FirstOrCreate finds the first record that matches the given attributes
	// or create a new one with those attributes if none was found.
	FirstOrCreate(dest any, conds ...any) error
	// FirstOr finds the first record that matches the given conditions or
	// execute the callback and return its result if no record is found.
	FirstOr(dest any, callback func() error) error
	// FirstOrFail finds the first record that matches the given conditions or throws an error.
	FirstOrFail(dest any) error
	// FirstOrNew finds the first record that matches the given conditions or
	// return a new instance of the model initialized with those attributes.
	FirstOrNew(dest any, attributes any, values ...any) error
	// ForceDelete forces delete records matching given conditions.
	ForceDelete(value any, conds ...any) (*Result, error)
	// Get retrieves all rows from the database.
	Get(dest any) error
	// Group specifies the group method on the query.
	Group(name string) Query
	// Having specifying HAVING conditions for the query.
	Having(query any, args ...any) Query
	// InRandomOrder specifies the order randomly.
	InRandomOrder() Query
	// Join specifying JOIN conditions for the query.
	Join(query string, args ...any) Query
	// Limit the number of records returned.
	Limit(limit int) Query
	// Load loads a relationship for the model.
	Load(dest any, relation string, args ...any) error
	// LoadMissing loads a relationship for the model that is not already loaded.
	LoadMissing(dest any, relation string, args ...any) error
	// LockForUpdate locks the selected rows in the table for updating.
	LockForUpdate() Query
	// Model sets the model instance to be queried.
	Model(value any) Query
	// Offset specifies the number of records to skip before starting to return the records.
	Offset(offset int) Query
	// Omit specifies columns that should be omitted from the query.
	Omit(columns ...string) Query
	// Order specifies the order in which the results should be returned.
	Order(value any) Query
	// OrderBy specifies the order should be ascending.
	OrderBy(column string, direction ...string) Query
	// OrderByDesc specifies the order should be descending.
	OrderByDesc(column string) Query
	// OrWhere add an "or where" clause to the query.
	OrWhere(query any, args ...any) Query
	// OrWhereIn adds an "or where column in" clause to the query.
	OrWhereIn(column string, values []any) Query
	// OrWhereNotIn adds an "or where column not in" clause to the query.
	OrWhereNotIn(column string, values []any) Query
	// OrWhereBetween adds an "or where column between x and y" clause to the query.
	OrWhereBetween(column string, x, y any) Query
	// OrWhereNotBetween adds an "or where column not between x and y" clause to the query.
	OrWhereNotBetween(column string, x, y any) Query
	// OrWhereNull adds a "or where column is null" clause to the query.
	OrWhereNull(column string) Query
	// Paginate the given query into a simple paginator.
	Paginate(page, limit int, dest any, total *int64) error
	// Pluck retrieves a single column from the database.
	Pluck(column string, dest any) error
	// Raw creates a raw query.
	Raw(sql string, values ...any) Query
	// Save updates value in a database
	Save(value any) error
	// SaveQuietly updates value in a database without firing events
	SaveQuietly(value any) error
	// Scan scans the query result and populates the destination object.
	Scan(dest any) error
	// Scopes applies one or more query scopes.
	Scopes(funcs ...func(Query) Query) Query
	// Select specifies fields that should be retrieved from the database.
	Select(query any, args ...any) Query
	// SharedLock locks the selected rows in the table.
	SharedLock() Query
	// Sum calculates the sum of a column's values and populates the destination object.
	Sum(column string, dest any) error
	// Table specifies the table for the query.
	Table(name string, args ...any) Query
	// ToSql returns the query as a SQL string.
	ToSql() ToSql
	// ToRawSql returns the query as a raw SQL string.
	ToRawSql() ToSql
	// Update updates records with the given column and values
	Update(column any, value ...any) (*Result, error)
	// UpdateOrCreate finds the first record that matches the given attributes
	// or create a new one with those attributes if none was found.
	UpdateOrCreate(dest any, attributes any, values any) error
	// Where add a "where" clause to the query.
	Where(query any, args ...any) Query
	// WhereIn adds a "where column in" clause to the query.
	WhereIn(column string, values []any) Query
	// WhereNotIn adds a "where column not in" clause to the query.
	WhereNotIn(column string, values []any) Query
	// WhereBetween adds a "where column between x and y" clause to the query.
	WhereBetween(column string, x, y any) Query
	// WhereNotBetween adds a "where column not between x and y" clause to the query.
	WhereNotBetween(column string, x, y any) Query
	// WhereNull adds a "where column is null" clause to the query.
	WhereNull(column string) Query
	// WhereNotNull adds a "where column is not null" clause to the query.
	WhereNotNull(column string) Query
	// WithoutEvents disables event firing for the query.
	WithoutEvents() Query
	// WithTrashed allows soft deleted models to be included in the results.
	WithTrashed() Query
	// With returns a new query instance with the given relationships eager loaded.
	With(query string, args ...any) Query
}

type Association interface {
	// Find finds records that match given conditions.
	Find(out any, conds ...any) error
	// Append appending a model to the association.
	Append(values ...any) error
	// Replace replaces the association with the given value.
	Replace(values ...any) error
	// Delete deletes the given value from the association.
	Delete(values ...any) error
	// Clear clears the association.
	Clear() error
	// Count returns the number of records in the association.
	Count() int64
}

type ConnectionModel interface {
	// Connection gets the connection name for the model.
	Connection() string
}

type Cursor interface {
	// Scan scans the current row into the given destination.
	Scan(value any) error
}

type Result struct {
	RowsAffected int64
}

type ToSql interface {
	Count() string
	Create(value any) string
	Delete(value any, conds ...any) string
	Find(dest any, conds ...any) string
	First(dest any) string
	Get(dest any) string
	Pluck(column string, dest any) string
	Save(value any) string
	Sum(column string, dest any) string
	Update(column any, value ...any) string
}
