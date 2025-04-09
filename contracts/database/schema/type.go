package schema

import "github.com/goravel/framework/support/enum"

type ColumnType = enum.Enum[string, string]

var (
	TypeBigInteger    = enum.NewEnum("BigInteger", "bigInteger")
	TypeBoolean       = enum.NewEnum("Boolean", "boolean")
	TypeChar          = enum.NewEnum("Char", "char")
	TypeDecimal       = enum.NewEnum("Decimal", "decimal")
	TypeDate          = enum.NewEnum("Date", "date")
	TypeDateTime      = enum.NewEnum("DateTime", "dateTime")
	TypeDateTimeTZ    = enum.NewEnum("DateTimeTZ", "dateTimeTz")
	TypeDouble        = enum.NewEnum("Double", "double")
	TypeEnum          = enum.NewEnum("Enum", "enum")
	TypeFloat         = enum.NewEnum("Float", "float")
	TypeInteger       = enum.NewEnum("Integer", "integer")
	TypeJson          = enum.NewEnum("JSON", "json")
	TypeJsonb         = enum.NewEnum("JSONB", "jsonb")
	TypeLongText      = enum.NewEnum("LongText", "longText")
	TypeMediumInteger = enum.NewEnum("MediumInteger", "mediumInteger")
	TypeMediumText    = enum.NewEnum("MediumText", "mediumText")
	TypeSmallInteger  = enum.NewEnum("SmallInteger", "smallInteger")
	TypeString        = enum.NewEnum("String", "string")
	TypeTime          = enum.NewEnum("Time", "time")
	TypeTimeTZ        = enum.NewEnum("TimeTZ", "timeTz")
	TypeTimestamp     = enum.NewEnum("Timestamp", "timestamp")
	TypeTimestampTZ   = enum.NewEnum("TimestampTZ", "timestampTz")
	TypeTinyInteger   = enum.NewEnum("TinyInteger", "tinyInteger")
	TypeTinyText      = enum.NewEnum("TinyText", "tinyText")
	TypeText          = enum.NewEnum("Text", "text")
)
