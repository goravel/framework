package grammars

import (
	"fmt"
	"strings"
)

type Wrap struct {
	tablePrefix string
}

func NewWrap(tablePrefix string) *Wrap {
	return &Wrap{
		tablePrefix: tablePrefix,
	}
}

func (r *Wrap) Column(column string) string {
	if strings.Contains(column, " as ") {
		return r.aliasedValue(column)
	}

	return r.Segments(strings.Split(column, "."))
}

func (r *Wrap) Columns(columns []string) string {
	for i, column := range columns {
		columns[i] = r.Column(column)
	}

	return strings.Join(columns, ", ")
}

func (r *Wrap) Quote(value string) string {
	if value == "" {
		return value
	}

	return fmt.Sprintf("'%s'", value)
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
		newTable := table[:lastDotIndex] + "." + r.tablePrefix
		if lastDotIndex+1 < len(table) {
			newTable += table[lastDotIndex+1:]
		}

		return r.Value(newTable)
	}

	return r.Value(r.tablePrefix + table)
}

func (r *Wrap) Value(value string) string {
	if value != "*" {
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
