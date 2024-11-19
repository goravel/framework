package processors

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

type DBIndex struct {
	Columns string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}
