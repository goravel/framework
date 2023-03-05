package validation

import (
	"errors"
	"reflect"
	"time"

	"github.com/gookit/validate"

	"github.com/goravel/framework/contracts/http"
	validatecontract "github.com/goravel/framework/contracts/validation"
)

type Validation struct {
	rules []validatecontract.Rule
}

func NewValidation() *Validation {
	return &Validation{
		rules: make([]validatecontract.Rule, 0),
	}
}

func (r *Validation) Make(data any, rules map[string]string, options ...validatecontract.Option) (validatecontract.Validator, error) {
	if data == nil {
		return nil, errors.New("data can't be empty")
	}
	if len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	var dataType reflect.Kind
	switch data := data.(type) {
	case map[string]any:
		if len(data) == 0 {
			return nil, errors.New("data can't be empty")
		}
		dataType = reflect.Map
	}

	val := reflect.ValueOf(data)
	indirectVal := reflect.Indirect(val)
	typ := indirectVal.Type()
	if indirectVal.Kind() == reflect.Struct && typ != reflect.TypeOf(time.Time{}) {
		dataType = reflect.Struct
	}

	var dataFace validate.DataFace
	switch dataType {
	case reflect.Map:
		dataFace = validate.FromMap(data.(map[string]any))
	case reflect.Struct:
		var err error
		dataFace, err = validate.FromStruct(data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("data must be map[string]any or struct")
	}

	options = append(options, Rules(rules), CustomRules(r.rules))
	generateOptions := GenerateOptions(options)
	if generateOptions["prepareForValidation"] != nil {
		if err := generateOptions["prepareForValidation"].(func(ctx http.Context, data validatecontract.Data) error)(nil, NewData(dataFace)); err != nil {
			return nil, err
		}
	}

	v := dataFace.Create()
	AppendOptions(v, generateOptions)

	return NewValidator(v, dataFace), nil
}

func (r *Validation) AddRules(rules []validatecontract.Rule) error {
	existRuleNames := r.existRuleNames()
	for _, rule := range rules {
		for _, existRuleName := range existRuleNames {
			if existRuleName == rule.Signature() {
				return errors.New("duplicate rule name: " + rule.Signature())
			}
		}
	}

	r.rules = append(r.rules, rules...)

	return nil
}

func (r *Validation) Rules() []validatecontract.Rule {
	return r.rules
}

func (r *Validation) existRuleNames() []string {
	rules := []string{
		"required",
		"required_if",
		"requiredIf",
		"required_unless",
		"requiredUnless",
		"required_with",
		"requiredWith",
		"required_with_all",
		"requiredWithAll",
		"required_without",
		"requiredWithout",
		"required_without_all",
		"requiredWithoutAll",
		"safe",
		"int",
		"integer",
		"isInt",
		"uint",
		"isUint",
		"bool",
		"isBool",
		"string",
		"isString",
		"float",
		"isFloat",
		"slice",
		"isSlice",
		"in",
		"enum",
		"not_in",
		"notIn",
		"contains",
		"not_contains",
		"notContains",
		"string_contains",
		"stringContains",
		"starts_with",
		"startsWith",
		"ends_with",
		"endsWith",
		"range",
		"between",
		"max",
		"lte",
		"min",
		"gte",
		"eq",
		"equal",
		"isEqual",
		"ne",
		"notEq",
		"notEqual",
		"lt",
		"lessThan",
		"gt",
		"greaterThan",
		"int_eq",
		"intEq",
		"intEqual",
		"len",
		"length",
		"min_len",
		"minLen",
		"minLength",
		"max_len",
		"maxLen",
		"maxLength",
		"email",
		"isEmail",
		"regex",
		"regexp",
		"arr",
		"list",
		"array",
		"isArray",
		"map",
		"isMap",
		"strings",
		"isStrings",
		"ints",
		"isInts",
		"eq_field",
		"eqField",
		"ne_field",
		"neField",
		"gte_field",
		"gtField",
		"gt_field",
		"gteField",
		"lt_field",
		"ltField",
		"lte_field",
		"lteField",
		"file",
		"isFile",
		"image",
		"isImage",
		"mime",
		"mimeType",
		"inMimeTypes",
		"date",
		"isDate",
		"gt_date",
		"gtDate",
		"afterDate",
		"lt_date",
		"ltDate",
		"beforeDate",
		"gte_date",
		"gteDate",
		"afterOrEqualDate",
		"lte_date",
		"lteDate",
		"beforeOrEqualDate",
		"hasWhitespace",
		"ascii",
		"ASCII",
		"isASCII",
		"alpha",
		"isAlpha",
		"alpha_num",
		"alphaNum",
		"isAlphaNum",
		"alpha_dash",
		"alphaDash",
		"isAlphaDash",
		"multi_byte",
		"multiByte",
		"isMultiByte",
		"base64",
		"isBase64",
		"dns_name",
		"dnsName",
		"DNSName",
		"isDNSName",
		"data_uri",
		"dataURI",
		"isDataURI",
		"empty",
		"isEmpty",
		"hex_color",
		"hexColor",
		"isHexColor",
		"hexadecimal",
		"isHexadecimal",
		"json",
		"JSON",
		"isJSON",
		"lat",
		"latitude",
		"isLatitude",
		"lon",
		"longitude",
		"isLongitude",
		"mac",
		"isMAC",
		"num",
		"number",
		"isNumber",
		"cn_mobile",
		"cnMobile",
		"isCnMobile",
		"printableASCII",
		"isPrintableASCII",
		"rgbColor",
		"RGBColor",
		"isRGBColor",
		"full_url",
		"fullUrl",
		"isFullURL",
		"url",
		"URL",
		"isURL",
		"ip",
		"IP",
		"isIP",
		"ipv4",
		"isIPv4",
		"ipv6",
		"isIPv6",
		"cidr",
		"CIDR",
		"isCIDR",
		"CIDRv4",
		"isCIDRv4",
		"CIDRv6",
		"isCIDRv6",
		"uuid",
		"isUUID",
		"uuid3",
		"isUUID3",
		"uuid4",
		"isUUID4",
		"uuid5",
		"isUUID5",
		"filePath",
		"isFilePath",
		"unixPath",
		"isUnixPath",
		"winPath",
		"isWinPath",
		"isbn10",
		"ISBN10",
		"isISBN10",
		"isbn13",
		"ISBN13",
		"isISBN13",
	}
	for _, rule := range r.rules {
		rules = append(rules, rule.Signature())
	}

	return rules
}
