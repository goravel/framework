package driver

type Processor interface {
	ProcessColumns(dbColumns []DBColumn) []Column
	ProcessForeignKeys(dbIndexes []DBForeignKey) []ForeignKey
	ProcessIndexes(dbIndexes []DBIndex) []Index
	ProcessTypes(types []Type) []Type
}

type DBColumn struct {
	Autoincrement bool
	Collation     string
	Comment       string
	Default       string
	Extra         string
	Length        int
	Name          string
	Nullable      string
	Places        int
	Precision     int
	Primary       bool
	Type          string
	TypeName      string
}

type DBForeignKey struct {
	Name           string
	Columns        string
	ForeignSchema  string
	ForeignTable   string
	ForeignColumns string
	OnUpdate       string
	OnDelete       string
}

type DBIndex struct {
	Columns string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}

type ForeignKey struct {
	Name           string
	Columns        []string
	ForeignSchema  string
	ForeignTable   string
	ForeignColumns []string
	OnUpdate       string
	OnDelete       string
}

type Index struct {
	Columns []string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}
