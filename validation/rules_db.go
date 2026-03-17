package validation

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
