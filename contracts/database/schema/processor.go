package schema

type Processor interface {
	// ProcessColumns Process the results of a columns query.
	ProcessColumns(columns []Column) []Column
}
