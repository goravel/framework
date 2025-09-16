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
	ArgumentFloat32   = ArgumentBase[float32]
	ArgumentFloat64   = ArgumentBase[float64]
	ArgumentInt       = ArgumentBase[int]
	ArgumentInt8      = ArgumentBase[int8]
	ArgumentInt16     = ArgumentBase[int16]
	ArgumentInt32     = ArgumentBase[int32]
	ArgumentInt64     = ArgumentBase[int64]
	ArgumentString    = ArgumentBase[string]
	ArgumentTimestamp = ArgumentBase[time.Time]
	ArgumentUint      = ArgumentBase[uint]
	ArgumentUint8     = ArgumentBase[uint8]
	ArgumentUint16    = ArgumentBase[uint16]
	ArgumentUint32    = ArgumentBase[uint32]
	ArgumentUint64    = ArgumentBase[uint64]

	ArgumentFloat32Slice   = ArgumentsBase[float32]
	ArgumentFloat64Slice   = ArgumentsBase[float64]
	ArgumentIntSlice       = ArgumentsBase[int]
	ArgumentInt8Slice      = ArgumentsBase[int8]
	ArgumentInt16Slice     = ArgumentsBase[int16]
	ArgumentInt32Slice     = ArgumentsBase[int32]
	ArgumentInt64Slice     = ArgumentsBase[int64]
	ArgumentStringSlice    = ArgumentsBase[string]
	ArgumentTimestampSlice = ArgumentsBase[time.Time]
	ArgumentUintSlice      = ArgumentsBase[uint]
	ArgumentUint8Slice     = ArgumentsBase[uint8]
	ArgumentUint16Slice    = ArgumentsBase[uint16]
	ArgumentUint32Slice    = ArgumentsBase[uint32]
	ArgumentUint64Slice    = ArgumentsBase[uint64]
)
