package schema

type ColumnDefinition struct {
	allowed       []string
	autoIncrement *bool
	change        *bool
	comment       *string
	def           *string
	length        *int
	name          *string
	nullable      *bool
	places        *int
	precision     *int
	total         *int
	ttype         *string
}

func (r *ColumnDefinition) Change() {
	*r.change = true
}

func (r *ColumnDefinition) GetAllowed() []string {
	return r.allowed
}

func (r *ColumnDefinition) GetAutoIncrement() bool {
	return *r.autoIncrement
}

func (r *ColumnDefinition) GetLength() int {
	return *r.length
}

func (r *ColumnDefinition) GetName() string {
	return *r.name
}

func (r *ColumnDefinition) GetPlaces() int {
	return *r.places
}

func (r *ColumnDefinition) GetPrecision() int {
	return *r.precision
}

func (r *ColumnDefinition) GetTotal() int {
	return *r.total
}

func (r *ColumnDefinition) GetType() string {
	return *r.ttype
}
