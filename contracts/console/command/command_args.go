package command

import "time"

type Argument interface {
	// Count of minimum occurrences
	MinOccurrences() int
	// Count of maximum occurrences
	MaxOccurrences() int
}

type ArgumentBase[T any] struct {
	Name     string // the name of this argument
	Value    T      // the default value of this argument
	Usage    string // the usage text to show
	Required bool   // if this argument is required
}

func (a ArgumentBase[T]) MinOccurrences() int {
	if a.Required {
		return 1
	} else {
		return 0
	}
}

func (a ArgumentBase[T]) MaxOccurrences() int {
	return 1
}

type ArgumentsBase[T any] struct {
	Name  string // the name of this argument
	Value T      // the default value of this argument
	Usage string // the usage text to show
	Min   int    // the min num of occurrences of this argument
	Max   int    // the max num of occurrences of this argument, set to -1 for unlimited
}

func (a ArgumentsBase[T]) MinOccurrences() int {
	return a.Min
}

func (a ArgumentsBase[T]) MaxOccurrences() int {
	return a.Max
}

type (
	Float32Argument   = ArgumentBase[float32]
	Float64Argument   = ArgumentBase[float64]
	IntArgument       = ArgumentBase[int]
	Int8Argument      = ArgumentBase[int8]
	Int16Argument     = ArgumentBase[int16]
	Int32Argument     = ArgumentBase[int32]
	Int64Argument     = ArgumentBase[int64]
	StringArgument    = ArgumentBase[string]
	TimestampArgument = ArgumentBase[time.Time]
	UintArgument      = ArgumentBase[uint]
	Uint8Argument     = ArgumentBase[uint8]
	Uint16Argument    = ArgumentBase[uint16]
	Uint32Argument    = ArgumentBase[uint32]
	Uint64Argument    = ArgumentBase[uint64]

	Float32SliceArgument   = ArgumentsBase[float32]
	Float64SliceArgument   = ArgumentsBase[float64]
	IntSliceArgument       = ArgumentsBase[int]
	Int8SliceArgument      = ArgumentsBase[int8]
	Int16SliceArgument     = ArgumentsBase[int16]
	Int32SliceArgument     = ArgumentsBase[int32]
	Int64SliceArgument     = ArgumentsBase[int64]
	StringSliceArgument    = ArgumentsBase[string]
	TimestampSliceArgument = ArgumentsBase[time.Time]
	UintSliceArgument      = ArgumentsBase[uint]
	Uint8SliceArgument     = ArgumentsBase[uint8]
	Uint16SliceArgument    = ArgumentsBase[uint16]
	Uint32SliceArgument    = ArgumentsBase[uint32]
	Uint64SliceArgument    = ArgumentsBase[uint64]
)
