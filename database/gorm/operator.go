package gorm

type Operator = string

const (
	gt  Operator = ">"
	gte Operator = ">="
	eq  Operator = "="
	lte Operator = "<="
	lt  Operator = "<"
)

func isAnyOperator(s any) (Operator, bool) {
	o, ok := s.(string)
	if !ok {
		return "", false
	}

	if isOperator(o) {
		return o, true
	}

	return "", false
}

func isOperator(s string) bool {
	switch s {
	case gt, gte, eq, lte, lt:
		return true
	default:
		return false
	}
}
