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

func (r *ColumnDefinition) GetAutoIncrement() (autoIncrement bool) {
	if r.autoIncrement != nil {
		return *r.autoIncrement
	}

	return
}

func (r *ColumnDefinition) GetLength() (length int) {
	if r.length != nil {
		return *r.length
	}

	return
}

func (r *ColumnDefinition) GetName() (name string) {
	if r.name != nil {
		return *r.name
	}

	return
}

func (r *ColumnDefinition) GetPlaces() (places int) {
	if r.places != nil {
		return *r.places
	}

	return
}

func (r *ColumnDefinition) GetPrecision() (precision int) {
	if r.precision != nil {
		return *r.precision
	}

	return
}

func (r *ColumnDefinition) GetTotal() (total int) {
	if r.total != nil {
		return *r.total
	}

	return
}

func (r *ColumnDefinition) GetType() (ttype string) {
	if r.ttype != nil {
		return *r.ttype
	}

	return
}
