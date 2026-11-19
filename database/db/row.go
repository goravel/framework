package db

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"gorm.io/gorm"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

// maxNameCacheEntries bounds the memoised Studly and Snake conversions. Column names come from the schema
// and field names from the model structs, so the set is small in practice; the bound only stops an
// application that builds names at runtime (aliases, pivot columns) from growing the caches forever.
const maxNameCacheEntries = 1024

var (
	// Type values compared by the decode hooks, resolved once instead of on every column and every row.
	typeString    = reflect.TypeFor[string]()
	typeTime      = reflect.TypeFor[time.Time]()
	typeDeletedAt = reflect.TypeFor[gorm.DeletedAt]()
	typeByteSlice = reflect.TypeFor[[]byte]()
	typeScanner   = reflect.TypeFor[interface{ Scan(any) error }]()

	// decodeHooks is composed once. Composing it per row rebuilt the same six hooks on every row.
	decodeHooks = mapstructure.ComposeDecodeHookFunc(
		ToStringHookFunc(), ToTimeHookFunc(), ToDeletedAtHookFunc(), ToScannerHookFunc(), ToSliceHookFunc(), ToMapHookFunc(),
	)

	studlyNames = newNameCache()
	snakeNames  = newNameCache()
)

// nameCache memoises a name conversion. It is read for every column and field pair of every row, so it
// uses a RWMutex and a plain map instead of a sync.Map.
type nameCache struct {
	mu sync.RWMutex
	m  map[string]string
}

func newNameCache() *nameCache {
	return &nameCache{m: make(map[string]string)}
}

func (c *nameCache) load(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	converted, ok := c.m[name]
	return converted, ok
}

func (c *nameCache) store(name, converted string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) >= maxNameCacheEntries {
		return
	}

	c.m[name] = converted
}

// studlyName returns name in StudlyCase. The conversion builds an x/text caser and the same names repeat
// across the rows of a result set, so the result is memoised.
func studlyName(name string) string {
	if converted, ok := studlyNames.load(name); ok {
		return converted
	}

	converted := str.Of(name).Studly().String()
	studlyNames.store(name, converted)

	return converted
}

// snakeName returns name in snake_case, memoised for the same reason as studlyName.
func snakeName(name string) string {
	if converted, ok := snakeNames.load(name); ok {
		return converted
	}

	converted := str.Of(name).Snake().String()
	snakeNames.store(name, converted)

	return converted
}

// matchName reports whether the column mapKey fills the struct field fieldName. The three original
// conditions are kept, with the one that needs no conversion first.
func matchName(mapKey, fieldName string) bool {
	return strings.EqualFold(mapKey, fieldName) ||
		studlyName(mapKey) == fieldName ||
		mapKey == snakeName(fieldName)
}

type Row struct {
	err error
	row map[string]any
}

func NewRow(row map[string]any, err error) *Row {
	return &Row{row: row, err: err}
}

func (r *Row) Err() error {
	return r.err
}

func (r *Row) Scan(value any) error {
	if r.err != nil {
		return r.err
	}

	msConfig := &mapstructure.DecoderConfig{
		DecodeHook: decodeHooks,
		Squash:     true,
		Result:     value,
		MatchName:  matchName,
	}

	decoder, err := mapstructure.NewDecoder(msConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(r.row)
}

// ToStringHookFunc is a hook function that converts []uint8 to string.
// Mysql returns []uint8 for String type when scanning the rows.
func ToStringHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != typeString {
			return data, nil
		}

		dataSlice, ok := data.([]uint8)
		if ok {
			return string(dataSlice), nil
		}

		return data, nil
	}
}

func ToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != typeTime {
			return data, nil
		}

		switch value := data.(type) {
		case string:
			return time.Parse(time.RFC3339, value)
		case float64:
			return time.Unix(0, int64(value)*int64(time.Millisecond)), nil
		case int64:
			return time.Unix(0, value*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
	}
}

func ToDeletedAtHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != typeDeletedAt {
			return data, nil
		}

		if f == typeTime {
			return gorm.DeletedAt{Time: data.(time.Time), Valid: true}, nil
		}

		if f.Kind() == reflect.String {
			value, ok := data.(string)
			if !ok {
				return data, nil
			}

			return gorm.DeletedAt{Time: carbon.Parse(value).StdTime(), Valid: true}, nil
		}

		return data, nil
	}
}

// ToScannerHookFunc is a hook function that handles types with custom Scan methods (sql.Scanner interface).
// This includes carbon types and other custom types implementing the Scan method.
func ToScannerHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		// Skip types that are handled by other specific hooks
		if t == typeTime || t == typeDeletedAt {
			return data, nil
		}

		// Skip if source and target are the same type
		if f == t {
			return data, nil
		}

		// Only process database types (string, []byte, time.Time)
		if f.Kind() != reflect.String && f != typeByteSlice && f != typeTime {
			return data, nil
		}

		// Create a pointer to the target type to check for Scan method
		targetPtr := reflect.PointerTo(t)
		if !targetPtr.Implements(typeScanner) {
			return data, nil
		}

		// Handle nil or empty data
		if data == nil {
			return reflect.Zero(t).Interface(), nil
		}
		if str, ok := data.(string); ok && str == "" {
			return reflect.Zero(t).Interface(), nil
		}

		// Create a new instance of the target type
		result := reflect.New(t)
		scanner := result.Interface().(interface{ Scan(any) error })

		// Convert string to []byte if needed (common for JSON fields from database)
		scanData := data
		if str, ok := data.(string); ok {
			scanData = []byte(str)
		}

		// Call the Scan method with the data
		if err := scanner.Scan(scanData); err != nil {
			return nil, err
		}

		return result.Elem().Interface(), nil
	}
}

// ToSliceHookFunc is a hook function that converts JSON string to slice.
func ToSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t.Kind() != reflect.Slice || f.Kind() != reflect.String {
			return data, nil
		}

		str, ok := data.(string)
		if !ok {
			return data, nil
		}

		// Return empty slice for empty string
		if str == "" {
			return reflect.MakeSlice(t, 0, 0).Interface(), nil
		}

		result := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(str), result); err != nil {
			return nil, err
		}

		return reflect.ValueOf(result).Elem().Interface(), nil
	}
}

// ToMapHookFunc is a hook function that converts JSON string to map.
func ToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t.Kind() != reflect.Map || f.Kind() != reflect.String {
			return data, nil
		}

		str, ok := data.(string)
		if !ok {
			return data, nil
		}

		// Return empty map for empty string
		if str == "" {
			return reflect.MakeMap(t).Interface(), nil
		}

		result := reflect.MakeMap(t).Interface()
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			return nil, err
		}

		return result, nil
	}
}
