package command

const (
	FlagTypeBool         = "bool"
	FlagTypeFloat64      = "float64"
	FlagTypeFloat64Slice = "float64_slice"
	FlagTypeInt          = "int"
	FlagTypeIntSlice     = "int_slice"
	FlagTypeInt64        = "int64"
	FlagTypeInt64Slice   = "int64_slice"
	FlagTypeString       = "string"
	FlagTypeStringSlice  = "string_slice"
)

type Extend struct {
	Category string
	Flags    []Flag
}

type Flag interface {
	// Type gets a flag type.
	Type() string
}

type BoolFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    bool
}

func (receiver *BoolFlag) Type() string {
	return FlagTypeBool
}

type Float64Flag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    float64
}

func (receiver *Float64Flag) Type() string {
	return FlagTypeFloat64
}

type Float64SliceFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    []float64
}

func (receiver *Float64SliceFlag) Type() string {
	return FlagTypeFloat64Slice
}

type IntFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    int
}

func (receiver *IntFlag) Type() string {
	return FlagTypeInt
}

type IntSliceFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    []int
}

func (receiver *IntSliceFlag) Type() string {
	return FlagTypeIntSlice
}

type Int64Flag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    int64
}

func (receiver *Int64Flag) Type() string {
	return FlagTypeInt64
}

type Int64SliceFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    []int64
}

func (receiver *Int64SliceFlag) Type() string {
	return FlagTypeInt64Slice
}

type StringFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    string
}

func (receiver *StringFlag) Type() string {
	return FlagTypeString
}

type StringSliceFlag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    []string
}

func (receiver *StringSliceFlag) Type() string {
	return FlagTypeStringSlice
}
