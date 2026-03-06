package validation

import (
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
)

// ruleExists validates that the given value exists in the specified database table.
// Syntax: exists:table,column1,column2,...
//   - table: required, supports "connection.table" format
//   - columns: optional, defaults to the current field name. Multiple columns are joined with OR.
func ruleExists(ctx *RuleContext) bool {
	if ormFacade == nil {
		return false
	}

	table, columns, connection := parseExistsParams(ctx)
	if table == "" {
		return false
	}

	query := getOrmQuery(ctx, connection).Table(table)

	if len(columns) == 1 {
		query = query.Where(columns[0], ctx.Value)
	} else {
		// Multiple columns: WHERE col1 = value OR col2 = value OR ...
		for i, col := range columns {
			if i == 0 {
				query = query.Where(col, ctx.Value)
			} else {
				query = query.OrWhere(col, ctx.Value)
			}
		}
	}

	exists, err := query.Exists()
	if err != nil {
		return false
	}
	return exists
}

// ruleUnique validates that the given value is unique in the specified database table.
// Syntax: unique:table,column,idColumn,except1,except2,...
//   - table: required, supports "connection.table" format
//   - column: optional, defaults to the current field name
//   - idColumn: optional, defaults to "id", the column to use for the except clause
//   - except values: optional, values to exclude from the unique check (for update scenarios)
func ruleUnique(ctx *RuleContext) bool {
	if ormFacade == nil {
		return false
	}

	table, column, connection := parseUniqueParams(ctx)
	if table == "" {
		return false
	}

	query := getOrmQuery(ctx, connection).Table(table).Where(column, ctx.Value)

	// Handle except (ignore specific records for updates)
	// Parameters: table, column, idColumn, except1, except2, ...
	if len(ctx.Parameters) >= 4 {
		idColumn := "id"
		if len(ctx.Parameters) >= 3 && ctx.Parameters[2] != "" {
			idColumn = ctx.Parameters[2]
		}

		var exceptValues []any
		for i := 3; i < len(ctx.Parameters); i++ {
			if ctx.Parameters[i] != "" {
				exceptValues = append(exceptValues, ctx.Parameters[i])
			}
		}

		if len(exceptValues) > 0 {
			query = query.WhereNotIn(idColumn, exceptValues)
		}
	}

	count, err := query.Count()
	if err != nil {
		return false
	}
	return count == 0
}

// parseExistsParams extracts table, columns, and connection from exists rule parameters.
// Supports "connection.table" format for specifying database connection.
// All parameters after the table name are treated as column names.
func parseExistsParams(ctx *RuleContext) (table string, columns []string, connection string) {
	if len(ctx.Parameters) == 0 {
		return "", []string{ctx.Attribute}, ""
	}

	table = ctx.Parameters[0]

	// Parse connection.table format
	if dotIdx := strings.Index(table, "."); dotIdx > 0 {
		connection = table[:dotIdx]
		table = table[dotIdx+1:]
	}

	// Collect all columns from parameters (starting at index 1)
	for i := 1; i < len(ctx.Parameters); i++ {
		if ctx.Parameters[i] != "" {
			columns = append(columns, ctx.Parameters[i])
		}
	}

	// Default column to field name if none specified
	if len(columns) == 0 {
		columns = []string{ctx.Attribute}
	}

	return table, columns, connection
}

// parseUniqueParams extracts table, column, and connection from unique rule parameters.
// Supports "connection.table" format for specifying database connection.
func parseUniqueParams(ctx *RuleContext) (table, column, connection string) {
	if len(ctx.Parameters) == 0 {
		return "", ctx.Attribute, ""
	}

	table = ctx.Parameters[0]

	// Parse connection.table format
	if dotIdx := strings.Index(table, "."); dotIdx > 0 {
		connection = table[:dotIdx]
		table = table[dotIdx+1:]
	}

	// Column defaults to field name
	column = ctx.Attribute
	if len(ctx.Parameters) >= 2 && ctx.Parameters[1] != "" {
		column = ctx.Parameters[1]
	}

	return table, column, connection
}

// getOrmQuery returns an ORM query, optionally with a specific connection.
func getOrmQuery(ctx *RuleContext, connection string) orm.Query {
	o := ormFacade.WithContext(ctx.Ctx)
	if connection != "" {
		o = o.Connection(connection)
	}
	return o.Query()
}
