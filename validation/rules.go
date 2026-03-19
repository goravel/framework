package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"mime/multipart"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cast"
)

// RuleContext provides context for rule evaluation.
type RuleContext struct {
	Ctx        context.Context
	Attribute  string                  // Current field name
	Value      any                     // Field value
	Parameters []string                // Rule parameters
	Data       *DataBag                // Full data set
	Rules      map[string][]ParsedRule // All field rules (for type inference)
}

// builtinRules maps rule names to their implementations.
var builtinRules = map[string]func(ctx *RuleContext) bool{
	// Existence
	"required":             ruleRequired,
	"required_if":          ruleRequiredIf,
	"required_unless":      ruleRequiredUnless,
	"required_with":        ruleRequiredWith,
	"required_with_all":    ruleRequiredWithAll,
	"required_without":     ruleRequiredWithout,
	"required_without_all": ruleRequiredWithoutAll,
	"required_if_accepted": ruleRequiredIfAccepted,
	"required_if_declined": ruleRequiredIfDeclined,
	"filled":               ruleFilled,
	"present":              rulePresent,
	"present_if":           rulePresentIf,
	"present_unless":       rulePresentUnless,
	"present_with":         rulePresentWith,
	"present_with_all":     rulePresentWithAll,
	"missing":              ruleMissing,
	"missing_if":           ruleMissingIf,
	"missing_unless":       ruleMissingUnless,
	"missing_with":         ruleMissingWith,
	"missing_with_all":     ruleMissingWithAll,

	// Accept/Decline
	"accepted":    ruleAccepted,
	"accepted_if": ruleAcceptedIf,
	"declined":    ruleDeclined,
	"declined_if": ruleDeclinedIf,

	// Prohibition
	"prohibited":             ruleProhibited,
	"prohibited_if":          ruleProhibitedIf,
	"prohibited_unless":      ruleProhibitedUnless,
	"prohibited_if_accepted": ruleProhibitedIfAccepted,
	"prohibited_if_declined": ruleProhibitedIfDeclined,
	"prohibits":              ruleProhibits,

	// Type
	"string":  ruleString,
	"integer": ruleInteger,
	"int":     ruleInteger, // Go alias
	"uint":    ruleUint,    // Go-specific
	"numeric": ruleNumeric,
	"boolean": ruleBoolean,
	"bool":    ruleBoolean, // Go alias
	"float":   ruleFloat,   // Go-specific
	"array":   ruleArray,
	"list":    ruleList,
	"slice":   ruleSlice, // Go alias for list
	"map":     ruleMap,   // Go-specific

	// Size
	"size":    ruleSize,
	"min":     ruleMin,
	"max":     ruleMax,
	"between": ruleBetween,
	"gt":      ruleGt,
	"gte":     ruleGte,
	"lt":      ruleLt,
	"lte":     ruleLte,

	// Numeric
	"digits":         ruleDigits,
	"digits_between": ruleDigitsBetween,
	"decimal":        ruleDecimal,
	"multiple_of":    ruleMultipleOf,
	"min_digits":     ruleMinDigits,
	"max_digits":     ruleMaxDigits,

	// String format
	"alpha":       ruleAlpha,
	"alpha_num":   ruleAlphaNum,
	"alpha_dash":  ruleAlphaDash,
	"ascii":       ruleAscii,
	"email":       ruleEmail,
	"url":         ruleUrl,
	"active_url":  ruleActiveUrl,
	"ip":          ruleIp,
	"ipv4":        ruleIpv4,
	"ipv6":        ruleIpv6,
	"mac_address": ruleMacAddress,
	"mac":         ruleMacAddress, // alias
	"json":        ruleJson,
	"uuid":        ruleUuid,
	"uuid3":       ruleUuid3,
	"uuid4":       ruleUuid4,
	"uuid5":       ruleUuid5,
	"ulid":        ruleUlid,
	"hex_color":   ruleHexColor,
	"regex":       ruleRegex,
	"not_regex":   ruleNotRegex,
	"lowercase":   ruleLowercase,
	"uppercase":   ruleUppercase,

	// String content
	"starts_with":       ruleStartsWith,
	"doesnt_start_with": ruleDoesntStartWith,
	"ends_with":         ruleEndsWith,
	"doesnt_end_with":   ruleDoesntEndWith,
	"contains":          ruleContains,
	"doesnt_contain":    ruleDoesntContain,
	"confirmed":         ruleConfirmed,

	// Comparison
	"same":          ruleSame,
	"different":     ruleDifferent,
	"eq":            ruleEq,
	"ne":            ruleNe,
	"in":            ruleIn,
	"not_in":        ruleNotIn,
	"in_array":      ruleInArray,
	"in_array_keys": ruleInArrayKeys,

	// Date
	"date":            ruleDate,
	"date_format":     ruleDateFormat,
	"date_equals":     ruleDateEquals,
	"before":          ruleBefore,
	"before_or_equal": ruleBeforeOrEqual,
	"after":           ruleAfter,
	"after_or_equal":  ruleAfterOrEqual,
	"timezone":        ruleTimezone,

	// Exclude (always return true; engine handles exclusion logic)
	"exclude":         ruleExclude,
	"exclude_if":      ruleExcludeIf,
	"exclude_unless":  ruleExcludeUnless,
	"exclude_with":    ruleExcludeWith,
	"exclude_without": ruleExcludeWithout,

	// File
	"file":       ruleFile,
	"image":      ruleImage,
	"mimes":      ruleMimes,
	"mimetypes":  ruleMimetypes,
	"extensions": ruleExtensions,
	"dimensions": ruleDimensions,
	"encoding":   ruleEncoding,

	// Control (always pass; handled by engine)
	"bail":      ruleBail,
	"nullable":  ruleNullable,
	"sometimes": ruleSometimes,

	// Other
	"distinct":            ruleDistinct,
	"required_array_keys": ruleRequiredArrayKeys,

	// Database
	"exists": ruleExists,
	"unique": ruleUnique,

	// Deprecated: use the new names instead, will be removed in the next version.
	"len":       ruleSize,          // use "size"
	"min_len":   ruleMin,           // use "min"
	"max_len":   ruleMax,           // use "max"
	"eq_field":  ruleSame,          // use "same"
	"ne_field":  ruleDifferent,     // use "different"
	"gt_field":  ruleGt,            // use "gt"
	"gte_field": ruleGte,           // use "gte"
	"lt_field":  ruleLt,            // use "lt"
	"lte_field": ruleLte,           // use "lte"
	"gt_date":   ruleAfter,         // use "after"
	"lt_date":   ruleBefore,        // use "before"
	"gte_date":  ruleAfterOrEqual,  // use "after_or_equal"
	"lte_date":  ruleBeforeOrEqual, // use "before_or_equal"
	"number":    ruleNumeric,       // use "numeric"
	"full_url":  ruleUrl,           // use "url"
}

// implicitRules are rules that run even when the field is missing or empty.
var implicitRules = map[string]bool{
	"required": true, "required_if": true, "required_unless": true,
	"required_with": true, "required_with_all": true,
	"required_without": true, "required_without_all": true,
	"required_if_accepted": true, "required_if_declined": true,
	"required_array_keys": true,
	"filled":              true,
	"present":             true, "present_if": true, "present_unless": true,
	"present_with": true, "present_with_all": true,
	"missing": true, "missing_if": true, "missing_unless": true,
	"missing_with": true, "missing_with_all": true,
	"accepted": true, "accepted_if": true,
	"declined": true, "declined_if": true,
	"prohibited": true, "prohibited_if": true, "prohibited_unless": true,
	"prohibited_if_accepted": true, "prohibited_if_declined": true,
	"prohibits": true,
}

// excludeRules are rules that may cause a field to be excluded from validated data.
var excludeRules = map[string]bool{
	"exclude": true, "exclude_if": true, "exclude_unless": true,
	"exclude_with": true, "exclude_without": true,
}

// numericRuleNames are rules that indicate a field should be treated as numeric for size rules.
var numericRuleNames = map[string]bool{
	"numeric": true, "integer": true, "decimal": true,
	"int": true, "uint": true, "float": true,
}

// ---- Existence Rules ----

func ruleRequired(ctx *RuleContext) bool {
	if ctx.Value == nil {
		return false
	}
	return !isValueEmpty(ctx.Value)
}

func ruleRequiredIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return ruleRequired(ctx)
	}
	return true
}

func ruleRequiredUnless(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if !matchesOtherValue(otherValue, comparisonValues) {
		return ruleRequired(ctx)
	}
	return true
}

func ruleRequiredWith(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if val, ok := ctx.Data.Get(field); ok && !isValueEmpty(val) {
			return ruleRequired(ctx)
		}
	}
	return true
}

func ruleRequiredWithAll(ctx *RuleContext) bool {
	allPresent := true
	for _, field := range ctx.Parameters {
		if val, ok := ctx.Data.Get(field); !ok || isValueEmpty(val) {
			allPresent = false
			break
		}
	}
	if allPresent {
		return ruleRequired(ctx)
	}
	return true
}

func ruleRequiredWithout(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if val, ok := ctx.Data.Get(field); !ok || isValueEmpty(val) {
			return ruleRequired(ctx)
		}
	}
	return true
}

func ruleRequiredWithoutAll(ctx *RuleContext) bool {
	nonePresent := true
	for _, field := range ctx.Parameters {
		if val, ok := ctx.Data.Get(field); ok && !isValueEmpty(val) {
			nonePresent = false
			break
		}
	}
	if nonePresent {
		return ruleRequired(ctx)
	}
	return true
}

func ruleRequiredIfAccepted(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return true
	}
	otherValue, _ := ctx.Data.Get(ctx.Parameters[0])
	if isAcceptedValue(otherValue) {
		return ruleRequired(ctx)
	}
	return true
}

func ruleRequiredIfDeclined(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return true
	}
	otherValue, _ := ctx.Data.Get(ctx.Parameters[0])
	if isDeclinedValue(otherValue) {
		return ruleRequired(ctx)
	}
	return true
}

func ruleFilled(ctx *RuleContext) bool {
	if !ctx.Data.Has(ctx.Attribute) {
		return true // Not present = ok for filled
	}
	return !isValueEmpty(ctx.Value)
}

func rulePresent(ctx *RuleContext) bool {
	return ctx.Data.Has(ctx.Attribute)
}

func rulePresentIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return ctx.Data.Has(ctx.Attribute)
	}
	return true
}

func rulePresentUnless(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if !matchesOtherValue(otherValue, comparisonValues) {
		return ctx.Data.Has(ctx.Attribute)
	}
	return true
}

func rulePresentWith(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if ctx.Data.Has(field) {
			return ctx.Data.Has(ctx.Attribute)
		}
	}
	return true
}

func rulePresentWithAll(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if !ctx.Data.Has(field) {
			return true
		}
	}
	return ctx.Data.Has(ctx.Attribute)
}

func ruleMissing(ctx *RuleContext) bool {
	return !ctx.Data.Has(ctx.Attribute)
}

func ruleMissingIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return !ctx.Data.Has(ctx.Attribute)
	}
	return true
}

func ruleMissingUnless(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if !matchesOtherValue(otherValue, comparisonValues) {
		return !ctx.Data.Has(ctx.Attribute)
	}
	return true
}

func ruleMissingWith(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if ctx.Data.Has(field) {
			return !ctx.Data.Has(ctx.Attribute)
		}
	}
	return true
}

func ruleMissingWithAll(ctx *RuleContext) bool {
	for _, field := range ctx.Parameters {
		if !ctx.Data.Has(field) {
			return true
		}
	}
	return !ctx.Data.Has(ctx.Attribute)
}

// ---- Accept/Decline Rules ----

func ruleAccepted(ctx *RuleContext) bool {
	return !isValueEmpty(ctx.Value) && isAcceptedValue(ctx.Value)
}

func ruleAcceptedIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return !isValueEmpty(ctx.Value) && isAcceptedValue(ctx.Value)
	}
	return true
}

func ruleDeclined(ctx *RuleContext) bool {
	return !isValueEmpty(ctx.Value) && isDeclinedValue(ctx.Value)
}

func ruleDeclinedIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return !isValueEmpty(ctx.Value) && isDeclinedValue(ctx.Value)
	}
	return true
}

// ---- Prohibition Rules ----

func ruleProhibited(ctx *RuleContext) bool {
	return isValueEmpty(ctx.Value)
}

func ruleProhibitedIf(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if matchesOtherValue(otherValue, comparisonValues) {
		return isValueEmpty(ctx.Value)
	}
	return true
}

func ruleProhibitedUnless(ctx *RuleContext) bool {
	otherValue, comparisonValues, _ := parseDependentValues(ctx)
	if !matchesOtherValue(otherValue, comparisonValues) {
		return isValueEmpty(ctx.Value)
	}
	return true
}

func ruleProhibitedIfAccepted(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return true
	}
	otherValue, _ := ctx.Data.Get(ctx.Parameters[0])
	if isAcceptedValue(otherValue) {
		return isValueEmpty(ctx.Value)
	}
	return true
}

func ruleProhibitedIfDeclined(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return true
	}
	otherValue, _ := ctx.Data.Get(ctx.Parameters[0])
	if isDeclinedValue(otherValue) {
		return isValueEmpty(ctx.Value)
	}
	return true
}

func ruleProhibits(ctx *RuleContext) bool {
	if isValueEmpty(ctx.Value) {
		return true
	}
	for _, field := range ctx.Parameters {
		if val, ok := ctx.Data.Get(field); ok && !isValueEmpty(val) {
			return false
		}
	}
	return true
}

// ---- Type Rules ----

func ruleString(ctx *RuleContext) bool {
	_, ok := ctx.Value.(string)
	return ok
}

func ruleInteger(ctx *RuleContext) bool {
	switch v := ctx.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return v == float32(int64(v))
	case float64:
		if v > float64(math.MaxInt64) || v < float64(math.MinInt64) {
			return false
		}
		return v == float64(int64(v))
	case json.Number:
		_, err := v.Int64()
		return err == nil
	case string:
		_, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return err == nil
	}
	return false
}

func ruleNumeric(ctx *RuleContext) bool {
	_, err := cast.ToFloat64E(ctx.Value)
	return err == nil
}

func ruleBoolean(ctx *RuleContext) bool {
	switch v := ctx.Value.(type) {
	case bool:
		return true
	case int:
		return v == 0 || v == 1
	case int64:
		return v == 0 || v == 1
	case float64:
		return v == 0 || v == 1
	case string:
		v = strings.TrimSpace(v)
		return v == "true" || v == "false" || v == "0" || v == "1" || v == "on" || v == "off" || v == "yes" || v == "no"
	}
	return false
}

func ruleFloat(ctx *RuleContext) bool {
	switch ctx.Value.(type) {
	case float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(ctx.Value.(string), 64)
		return err == nil
	}
	return false
}

func ruleArray(ctx *RuleContext) bool {
	if ctx.Value == nil {
		return false
	}
	rv := reflect.ValueOf(ctx.Value)
	kind := rv.Kind()
	return kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map
}

func ruleList(ctx *RuleContext) bool {
	// A list is an array with sequential integer keys (a Go slice)
	if ctx.Value == nil {
		return false
	}
	kind := reflect.ValueOf(ctx.Value).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func ruleSlice(ctx *RuleContext) bool {
	// Alias for list — validates value is a Go slice
	return ruleList(ctx)
}

func ruleMap(ctx *RuleContext) bool {
	// Validates value is a Go map
	if ctx.Value == nil {
		return false
	}
	return reflect.ValueOf(ctx.Value).Kind() == reflect.Map
}

// ---- Size Rules ----

func ruleSize(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	expected, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	return size == expected
}

func ruleMin(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	minV, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	return size >= minV
}

func ruleMax(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	maxV, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	return size <= maxV
}

func ruleBetween(ctx *RuleContext) bool {
	if len(ctx.Parameters) < 2 {
		return false
	}
	minV, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	maxV, err := strconv.ParseFloat(ctx.Parameters[1], 64)
	if err != nil {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	return size >= minV && size <= maxV
}

func ruleGt(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	// Try as field reference
	if otherVal, exists := ctx.Data.Get(ctx.Parameters[0]); exists {
		otherType := getAttributeType(ctx.Parameters[0], otherVal, ctx.Rules)
		otherSize, ok := getSize(otherVal, otherType)
		if ok {
			return size > otherSize
		}
	}
	// Try as numeric value
	threshold, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	return size > threshold
}

func ruleGte(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	if otherVal, exists := ctx.Data.Get(ctx.Parameters[0]); exists {
		otherType := getAttributeType(ctx.Parameters[0], otherVal, ctx.Rules)
		otherSize, ok := getSize(otherVal, otherType)
		if ok {
			return size >= otherSize
		}
	}
	threshold, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	return size >= threshold
}

func ruleLt(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	if otherVal, exists := ctx.Data.Get(ctx.Parameters[0]); exists {
		otherType := getAttributeType(ctx.Parameters[0], otherVal, ctx.Rules)
		otherSize, ok := getSize(otherVal, otherType)
		if ok {
			return size < otherSize
		}
	}
	threshold, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	return size < threshold
}

func ruleLte(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	attrType := getAttributeType(ctx.Attribute, ctx.Value, ctx.Rules)
	size, ok := getSize(ctx.Value, attrType)
	if !ok {
		return false
	}
	if otherVal, exists := ctx.Data.Get(ctx.Parameters[0]); exists {
		otherType := getAttributeType(ctx.Parameters[0], otherVal, ctx.Rules)
		otherSize, ok := getSize(otherVal, otherType)
		if ok {
			return size <= otherSize
		}
	}
	threshold, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil {
		return false
	}
	return size <= threshold
}

// ---- Numeric Rules ----

func ruleDigits(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	s = strings.TrimSpace(s)
	expected, err := strconv.Atoi(ctx.Parameters[0])
	if err != nil {
		return false
	}
	// Must be all digits
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) == expected
}

func ruleDigitsBetween(ctx *RuleContext) bool {
	if len(ctx.Parameters) < 2 {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	s = strings.TrimSpace(s)
	minV, err := strconv.Atoi(ctx.Parameters[0])
	if err != nil {
		return false
	}
	maxV, err := strconv.Atoi(ctx.Parameters[1])
	if err != nil {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	l := len(s)
	return l >= minV && l <= maxV
}

func ruleDecimal(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	s = strings.TrimSpace(s)

	// Parse expected decimal places
	minPlaces, err := strconv.Atoi(ctx.Parameters[0])
	if err != nil {
		return false
	}
	maxPlaces := minPlaces
	if len(ctx.Parameters) > 1 {
		maxPlaces, err = strconv.Atoi(ctx.Parameters[1])
		if err != nil {
			return false
		}
	}

	// Must be numeric
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return false
	}

	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return minPlaces == 0
	}
	decimalLen := len(parts[1])
	return decimalLen >= minPlaces && decimalLen <= maxPlaces
}

func ruleMultipleOf(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	val, err := cast.ToFloat64E(ctx.Value)
	if err != nil {
		return false
	}
	divisor, err := strconv.ParseFloat(ctx.Parameters[0], 64)
	if err != nil || divisor == 0 {
		return false
	}
	remainder := math.Mod(val, divisor)
	epsilon := 1e-9
	return math.Abs(remainder) < epsilon || math.Abs(remainder-divisor) < epsilon
}

func ruleMinDigits(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	s = strings.TrimSpace(s)
	// Remove non-digit characters for counting
	digitCount := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	minV, err := strconv.Atoi(ctx.Parameters[0])
	if err != nil {
		return false
	}
	// Must be numeric
	if _, err = strconv.ParseFloat(s, 64); err != nil {
		return false
	}
	return digitCount >= minV
}

func ruleMaxDigits(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	s = strings.TrimSpace(s)
	digitCount := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}
	maxV, err := strconv.Atoi(ctx.Parameters[0])
	if err != nil {
		return false
	}
	if _, err = strconv.ParseFloat(s, 64); err != nil {
		return false
	}
	return digitCount <= maxV
}

// ---- String Format Rules ----

func ruleAlpha(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return len(s) > 0
}

func ruleAlphaNum(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func ruleAlphaDash(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return len(s) > 0
}

func ruleAscii(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func ruleEmail(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	// mail.ParseAddress accepts "Name <email>" format, but validation
	// should only accept bare email addresses.
	return addr.Address == s
}

func ruleUrl(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func ruleActiveUrl(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return false
	}
	resolver := net.Resolver{}
	_, err = resolver.LookupHost(ctx.Ctx, u.Hostname())
	return err == nil
}

func ruleIp(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return net.ParseIP(s) != nil
}

func ruleIpv4(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func ruleIpv6(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

func ruleMacAddress(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return macRegex.MatchString(s)
}

func ruleJson(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ruleUuid(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return uuidRegex.MatchString(s)
}

var uuid3Regex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-3[0-9a-fA-F]{3}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ruleUuid3(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return uuid3Regex.MatchString(s)
}

var uuid4Regex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func ruleUuid4(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return uuid4Regex.MatchString(s)
}

var uuid5Regex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-5[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func ruleUuid5(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return uuid5Regex.MatchString(s)
}

var ulidRegex = regexp.MustCompile(`^[0-9A-HJ-KM-NP-TV-Za-hj-km-np-tv-z]{26}$`)

func ruleUlid(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return ulidRegex.MatchString(s)
}

var hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

func ruleHexColor(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return hexColorRegex.MatchString(s)
}

func ruleRegex(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	if len(ctx.Parameters) == 0 || ctx.Parameters[0] == "" {
		return false
	}
	pattern := ctx.Parameters[0]
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func ruleNotRegex(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	if len(ctx.Parameters) == 0 {
		return false
	}
	pattern := ctx.Parameters[0]
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return !re.MatchString(s)
}

func ruleLowercase(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return s == strings.ToLower(s)
}

func ruleUppercase(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	return s == strings.ToUpper(s)
}

// ---- String Content Rules ----

func ruleStartsWith(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, prefix := range ctx.Parameters {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func ruleDoesntStartWith(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, prefix := range ctx.Parameters {
		if strings.HasPrefix(s, prefix) {
			return false
		}
	}
	return true
}

func ruleEndsWith(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, suffix := range ctx.Parameters {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

func ruleDoesntEndWith(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, suffix := range ctx.Parameters {
		if strings.HasSuffix(s, suffix) {
			return false
		}
	}
	return true
}

func ruleContains(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, substr := range ctx.Parameters {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

func ruleDoesntContain(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, substr := range ctx.Parameters {
		if strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

func ruleConfirmed(ctx *RuleContext) bool {
	confirmField := ctx.Attribute + "_confirmation"
	confirmVal, ok := ctx.Data.Get(confirmField)
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", ctx.Value) == fmt.Sprintf("%v", confirmVal)
}

// ---- Comparison Rules ----

func ruleSame(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	otherVal, ok := ctx.Data.Get(ctx.Parameters[0])
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", ctx.Value) == fmt.Sprintf("%v", otherVal)
}

func ruleDifferent(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	otherVal, ok := ctx.Data.Get(ctx.Parameters[0])
	if !ok {
		return true
	}
	return fmt.Sprintf("%v", ctx.Value) != fmt.Sprintf("%v", otherVal)
}

func ruleEq(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	return fmt.Sprintf("%v", ctx.Value) == ctx.Parameters[0]
}

func ruleNe(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	return fmt.Sprintf("%v", ctx.Value) != ctx.Parameters[0]
}

func ruleUint(ctx *RuleContext) bool {
	if !ruleInteger(ctx) {
		return false
	}
	return cast.ToInt64(ctx.Value) >= 0
}

func ruleIn(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, allowed := range ctx.Parameters {
		if s == allowed {
			return true
		}
	}
	return false
}

func ruleNotIn(ctx *RuleContext) bool {
	s := fmt.Sprintf("%v", ctx.Value)
	for _, disallowed := range ctx.Parameters {
		if s == disallowed {
			return false
		}
	}
	return true
}

func ruleInArray(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	otherField := ctx.Parameters[0]
	otherVal, ok := ctx.Data.Get(otherField)
	if !ok {
		return false
	}
	s := fmt.Sprintf("%v", ctx.Value)
	switch arr := otherVal.(type) {
	case []any:
		for _, item := range arr {
			if fmt.Sprintf("%v", item) == s {
				return true
			}
		}
	case []string:
		for _, item := range arr {
			if item == s {
				return true
			}
		}
	}
	return false
}

// ---- Date Rules ----

func ruleDate(ctx *RuleContext) bool {
	_, ok := parseDate(ctx.Value)
	return ok
}

func ruleDateFormat(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	// Parameters[0] is a Go time layout (e.g., "2006-01-02 15:04:05")
	_, err := time.Parse(ctx.Parameters[0], s)
	return err == nil
}

func ruleDateEquals(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	valDate, ok := parseDate(ctx.Value)
	if !ok {
		return false
	}
	otherDate, ok := parseDateValue(ctx.Parameters[0], ctx.Data)
	if !ok {
		return false
	}
	return valDate.Equal(otherDate)
}

func ruleBefore(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	valDate, ok := parseDate(ctx.Value)
	if !ok {
		return false
	}
	otherDate, ok := parseDateValue(ctx.Parameters[0], ctx.Data)
	if !ok {
		return false
	}
	return valDate.Before(otherDate)
}

func ruleBeforeOrEqual(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	valDate, ok := parseDate(ctx.Value)
	if !ok {
		return false
	}
	otherDate, ok := parseDateValue(ctx.Parameters[0], ctx.Data)
	if !ok {
		return false
	}
	return valDate.Before(otherDate) || valDate.Equal(otherDate)
}

func ruleAfter(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	valDate, ok := parseDate(ctx.Value)
	if !ok {
		return false
	}
	otherDate, ok := parseDateValue(ctx.Parameters[0], ctx.Data)
	if !ok {
		return false
	}
	return valDate.After(otherDate)
}

func ruleAfterOrEqual(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	valDate, ok := parseDate(ctx.Value)
	if !ok {
		return false
	}
	otherDate, ok := parseDateValue(ctx.Parameters[0], ctx.Data)
	if !ok {
		return false
	}
	return valDate.After(otherDate) || valDate.Equal(otherDate)
}

// ---- Exclude Rules ----
// These always return true; the engine handles the actual exclusion logic.

func ruleExclude(ctx *RuleContext) bool {
	return true
}

func ruleExcludeIf(ctx *RuleContext) bool {
	return true
}

func ruleExcludeUnless(ctx *RuleContext) bool {
	return true
}

func ruleExcludeWith(ctx *RuleContext) bool {
	return true
}

func ruleExcludeWithout(ctx *RuleContext) bool {
	return true
}

// ---- File Rules ----

func ruleFile(ctx *RuleContext) bool {
	if _, ok := ctx.Value.(*multipart.FileHeader); ok {
		return true
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		return len(fhs) > 0
	}
	return false
}

func ruleImage(ctx *RuleContext) bool {
	check := func(fh *multipart.FileHeader) bool {
		mtype, err := detectMIME(fh)
		if err != nil {
			return false
		}
		imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}
		for _, t := range imageTypes {
			if mtype.Is(t) {
				return true
			}
		}
		return false
	}

	if fh, ok := ctx.Value.(*multipart.FileHeader); ok {
		return check(fh)
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if !check(fh) {
				return false
			}
		}
		return len(fhs) > 0
	}
	return false
}

func ruleMimes(ctx *RuleContext) bool {
	check := func(fh *multipart.FileHeader) bool {
		mtype, err := detectMIME(fh)
		if err != nil {
			return false
		}
		ext := strings.TrimPrefix(mtype.Extension(), ".")
		for _, allowed := range ctx.Parameters {
			if strings.EqualFold(ext, allowed) {
				return true
			}
		}
		return false
	}

	if fh, ok := ctx.Value.(*multipart.FileHeader); ok {
		return check(fh)
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if !check(fh) {
				return false
			}
		}
		return len(fhs) > 0
	}
	return false
}

func ruleMimetypes(ctx *RuleContext) bool {
	check := func(fh *multipart.FileHeader) bool {
		mtype, err := detectMIME(fh)
		if err != nil {
			return false
		}
		for _, allowed := range ctx.Parameters {
			if mtype.Is(allowed) {
				return true
			}
		}
		return false
	}

	if fh, ok := ctx.Value.(*multipart.FileHeader); ok {
		return check(fh)
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if !check(fh) {
				return false
			}
		}
		return len(fhs) > 0
	}
	return false
}

func ruleExtensions(ctx *RuleContext) bool {
	check := func(fh *multipart.FileHeader) bool {
		ext := strings.ToLower(getFileExtension(fh.Filename))
		for _, allowed := range ctx.Parameters {
			if ext == strings.ToLower(allowed) {
				return true
			}
		}
		return false
	}

	if fh, ok := ctx.Value.(*multipart.FileHeader); ok {
		return check(fh)
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if !check(fh) {
				return false
			}
		}
		return len(fhs) > 0
	}
	return false
}

func ruleDimensions(ctx *RuleContext) bool {
	// Parse named parameters: "min_width=100,max_width=500,width=200,height=200,ratio=3/2"
	constraints := make(map[string]string, len(ctx.Parameters))
	for _, p := range ctx.Parameters {
		if k, v, found := strings.Cut(p, "="); found {
			constraints[k] = v
		}
	}

	check := func(fh *multipart.FileHeader) bool {
		f, err := fh.Open()
		if err != nil {
			return false
		}
		defer func(f multipart.File) { _ = f.Close() }(f)

		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return false
		}

		width, height := cfg.Width, cfg.Height

		if v, ok := constraints["width"]; ok {
			if w, err := strconv.Atoi(v); err == nil && width != w {
				return false
			}
		}
		if v, ok := constraints["height"]; ok {
			if h, err := strconv.Atoi(v); err == nil && height != h {
				return false
			}
		}
		if v, ok := constraints["min_width"]; ok {
			if mw, err := strconv.Atoi(v); err == nil && width < mw {
				return false
			}
		}
		if v, ok := constraints["max_width"]; ok {
			if mw, err := strconv.Atoi(v); err == nil && width > mw {
				return false
			}
		}
		if v, ok := constraints["min_height"]; ok {
			if mh, err := strconv.Atoi(v); err == nil && height < mh {
				return false
			}
		}
		if v, ok := constraints["max_height"]; ok {
			if mh, err := strconv.Atoi(v); err == nil && height > mh {
				return false
			}
		}

		if v, ok := constraints["ratio"]; ok {
			var targetRatio float64
			if num, den, found := strings.Cut(v, "/"); found {
				n, err1 := strconv.ParseFloat(num, 64)
				d, err2 := strconv.ParseFloat(den, 64)
				if err1 != nil || err2 != nil || d == 0 {
					return false
				}
				targetRatio = n / d
			} else {
				r, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return false
				}
				targetRatio = r
			}
			actualRatio := float64(width) / float64(height)
			if math.Abs(actualRatio-targetRatio) > 0.01 {
				return false
			}
		}

		return true
	}

	if fh, ok := ctx.Value.(*multipart.FileHeader); ok {
		return check(fh)
	}
	if fhs, ok := ctx.Value.([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if !check(fh) {
				return false
			}
		}
		return len(fhs) > 0
	}
	return false
}

// ---- Control Rules ----

func ruleBail(_ *RuleContext) bool {
	return true
}

func ruleNullable(_ *RuleContext) bool {
	return true
}

func ruleSometimes(_ *RuleContext) bool {
	return true
}

// ---- Other Rules ----

func ruleDistinct(ctx *RuleContext) bool {
	// Distinct checks for duplicate values in array fields.
	// The engine handles tracking unique values across wildcard-expanded fields.
	return true
}

func ruleRequiredArrayKeys(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	if ctx.Value == nil {
		return false
	}
	rv := reflect.ValueOf(ctx.Value)
	if rv.Kind() != reflect.Map {
		return false
	}
	keys := rv.MapKeys()
	for _, param := range ctx.Parameters {
		found := false
		for _, k := range keys {
			if fmt.Sprintf("%v", k.Interface()) == param {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func ruleInArrayKeys(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}
	if ctx.Value == nil {
		return false
	}
	rv := reflect.ValueOf(ctx.Value)
	if rv.Kind() != reflect.Map {
		return false
	}
	keys := rv.MapKeys()
	for _, param := range ctx.Parameters {
		for _, k := range keys {
			if fmt.Sprintf("%v", k.Interface()) == param {
				return true
			}
		}
	}
	return false
}

func ruleTimezone(ctx *RuleContext) bool {
	s, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	_, err := time.LoadLocation(s)
	return err == nil
}

func ruleEncoding(ctx *RuleContext) bool {
	if len(ctx.Parameters) == 0 {
		return false
	}

	var data []byte
	switch v := ctx.Value.(type) {
	case string:
		data = []byte(v)
	case *multipart.FileHeader:
		f, err := v.Open()
		if err != nil {
			return false
		}
		defer func(f multipart.File) { _ = f.Close() }(f)
		data, err = io.ReadAll(f)
		if err != nil {
			return false
		}
	default:
		return false
	}

	enc := strings.ToLower(ctx.Parameters[0])
	switch enc {
	case "utf-8", "utf8":
		return utf8.Valid(data)
	case "ascii", "us-ascii":
		for _, b := range data {
			if b > 127 {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// ---- Database Rules ----

// ruleExists validates that the given value exists in the specified database table.
// Syntax: exists:table,column1,column2,...
//   - table: required, supports "connection.table" format
//   - columns: optional, defaults to the current field name. Multiple columns are joined with OR.
func ruleExists(ctx *RuleContext) bool {
	if ormFacade == nil || len(ctx.Parameters) == 0 {
		return false
	}

	table := ctx.Parameters[0]
	var connection string
	if dotIdx := strings.Index(table, "."); dotIdx > 0 {
		connection = table[:dotIdx]
		table = table[dotIdx+1:]
	}
	if table == "" {
		return false
	}

	var columns []string
	for i := 1; i < len(ctx.Parameters); i++ {
		if ctx.Parameters[i] != "" {
			columns = append(columns, ctx.Parameters[i])
		}
	}
	if len(columns) == 0 {
		columns = []string{ctx.Attribute}
	}

	o := ormFacade.WithContext(ctx.Ctx)
	if connection != "" {
		o = o.Connection(connection)
	}
	query := o.Query().Table(table)

	if len(columns) == 1 {
		query = query.Where(columns[0], ctx.Value)
	} else {
		for i, col := range columns {
			if i == 0 {
				query = query.Where(col, ctx.Value)
			} else {
				query = query.OrWhere(col, ctx.Value)
			}
		}
	}

	exists, err := query.Exists()
	if err != nil {
		return false
	}
	return exists
}

// ruleUnique validates that the given value is unique in the specified database table.
// Syntax: unique:table,column,idColumn,except1,except2,...
//   - table: required, supports "connection.table" format
//   - column: optional, defaults to the current field name
//   - idColumn: optional, defaults to "id", the column to use for the except clause
//   - except values: optional, values to exclude from the unique check (for update scenarios)
func ruleUnique(ctx *RuleContext) bool {
	if ormFacade == nil || len(ctx.Parameters) == 0 {
		return false
	}

	table := ctx.Parameters[0]
	var connection string
	if dotIdx := strings.Index(table, "."); dotIdx > 0 {
		connection = table[:dotIdx]
		table = table[dotIdx+1:]
	}
	if table == "" {
		return false
	}

	column := ctx.Attribute
	if len(ctx.Parameters) >= 2 && ctx.Parameters[1] != "" {
		column = ctx.Parameters[1]
	}

	o := ormFacade.WithContext(ctx.Ctx)
	if connection != "" {
		o = o.Connection(connection)
	}
	query := o.Query().Table(table).Where(column, ctx.Value)

	// Handle except (ignore specific records for updates)
	// Parameters: table, column, idColumn, except1, except2, ...
	if len(ctx.Parameters) >= 4 {
		idColumn := "id"
		if ctx.Parameters[2] != "" {
			idColumn = ctx.Parameters[2]
		}

		var exceptValues []any
		for i := 3; i < len(ctx.Parameters); i++ {
			if ctx.Parameters[i] != "" {
				exceptValues = append(exceptValues, ctx.Parameters[i])
			}
		}

		if len(exceptValues) > 0 {
			query = query.WhereNotIn(idColumn, exceptValues)
		}
	}

	count, err := query.Count()
	if err != nil {
		return false
	}
	return count == 0
}
