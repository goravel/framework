package grammars

import (
	"fmt"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/support/collect"
)

type Wrap struct {
	driver      contractsdatabase.Driver
	tablePrefix string
}

func NewWrap(driver contractsdatabase.Driver, tablePrefix string) *Wrap {
	return &Wrap{
		driver:      driver,
		tablePrefix: tablePrefix,
	}
}

func (r *Wrap) Column(column string) string {
	if strings.Contains(column, " as ") {
		return r.aliasedValue(column)
	}

	return r.Segments(strings.Split(column, "."))
}

func (r *Wrap) Columns(columns []string) []string {
	for i, column := range columns {
		columns[i] = r.Column(column)
	}

	return columns
}

func (r *Wrap) Columnize(columns []string) string {
	columns = r.Columns(columns)

	return strings.Join(columns, ", ")
}

func (r *Wrap) Quote(value string) string {
	if value == "" {
		return value
	}

	return fmt.Sprintf("'%s'", value)
}

func (r *Wrap) Quotes(value []string) []string {
	return collect.Map(value, func(v string, _ int) string {
		if r.driver == contractsdatabase.DriverSqlserver {
			return "N" + r.Quote(v)
		}
		return r.Quote(v)
	})
}

func (r *Wrap) Segments(segments []string) string {
	for i, segment := range segments {
		if i == 0 && len(segments) > 1 {
			segments[i] = r.Table(segment)
		} else {
			segments[i] = r.Value(segment)
		}
	}

	return strings.Join(segments, ".")
}

func (r *Wrap) Table(table string) string {
	if strings.Contains(table, " as ") {
		return r.aliasedTable(table)
	}
	if strings.Contains(table, ".") {
		lastDotIndex := strings.LastIndex(table, ".")
		newTable := table[:lastDotIndex] + "." + r.tablePrefix + table[lastDotIndex+1:]

		return r.Value(newTable)
	}

	return r.Value(r.tablePrefix + table)
}

func (r *Wrap) Value(value string) string {
	if value != "*" {
		if r.driver == contractsdatabase.DriverMysql {
			return "`" + strings.ReplaceAll(value, "`", "``") + "`"
		}
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	}

	return value
}

func (r *Wrap) aliasedTable(table string) string {
	segments := strings.Split(table, " as ")

	return r.Table(segments[0]) + " as " + r.Value(r.tablePrefix+segments[1])
}

func (r *Wrap) aliasedValue(value string) string {
	segments := strings.Split(value, " as ")

	return r.Column(segments[0]) + " as " + r.Value(r.tablePrefix+segments[1])
}
