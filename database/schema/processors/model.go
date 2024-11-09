package processors

type DBIndex struct {
	Columns string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}
