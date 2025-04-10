package schema

import "github.com/goravel/framework/support/enum"

type ColumnType = enum.Enum[string, string]

var (
	TypeBigInteger    = enum.New("BigInteger", "bigInteger")
	TypeBoolean       = enum.New("Boolean", "boolean")
	TypeChar          = enum.New("Char", "char")
	TypeDecimal       = enum.New("Decimal", "decimal")
	TypeDate          = enum.New("Date", "date")
	TypeDateTime      = enum.New("DateTime", "dateTime")
	TypeDateTimeTZ    = enum.New("DateTimeTZ", "dateTimeTz")
	TypeDouble        = enum.New("Double", "double")
	TypeEnum          = enum.New("Enum", "enum")
	TypeFloat         = enum.New("Float", "float")
	TypeInteger       = enum.New("Integer", "integer")
	TypeJson          = enum.New("JSON", "json")
	TypeJsonb         = enum.New("JSONB", "jsonb")
	TypeLongText      = enum.New("LongText", "longText")
	TypeMediumInteger = enum.New("MediumInteger", "mediumInteger")
	TypeMediumText    = enum.New("MediumText", "mediumText")
	TypeSmallInteger  = enum.New("SmallInteger", "smallInteger")
	TypeString        = enum.New("String", "string")
	TypeTime          = enum.New("Time", "time")
	TypeTimeTZ        = enum.New("TimeTZ", "timeTz")
	TypeTimestamp     = enum.New("Timestamp", "timestamp")
	TypeTimestampTZ   = enum.New("TimestampTZ", "timestampTz")
	TypeTinyInteger   = enum.New("TinyInteger", "tinyInteger")
	TypeTinyText      = enum.New("TinyText", "tinyText")
	TypeText          = enum.New("Text", "text")
)
