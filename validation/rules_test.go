package validation

import (
	"bytes"
	"context"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

type RulesTestSuite struct {
	suite.Suite
	validation *Validation
}

func (s *RulesTestSuite) SetupTest() {
	s.validation = NewValidation()
}

func (s *RulesTestSuite) makeValidator(data map[string]any, rules map[string]any, options ...contractsvalidation.Option) contractsvalidation.Validator {
	validator, err := s.validation.Make(context.Background(), data, rules, options...)
	s.Require().NoError(err)
	return validator
}

func TestRulesTestSuite(t *testing.T) {
	suite.Run(t, new(RulesTestSuite))
}

// ===== 1. Existence Rules =====

func (s *RulesTestSuite) TestRequired() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_with_value", map[string]any{"name": "goravel"}, map[string]any{"name": "required"}, false},
		{"pass_with_int", map[string]any{"age": 18}, map[string]any{"age": "required"}, false},
		{"pass_with_int_zero", map[string]any{"age": 0}, map[string]any{"age": "required"}, false},
		{"pass_with_bool_true", map[string]any{"ok": true}, map[string]any{"ok": "required"}, false},
		{"pass_with_bool_false", map[string]any{"ok": false}, map[string]any{"ok": "required"}, false},
		{"pass_with_float", map[string]any{"score": 3.14}, map[string]any{"score": "required"}, false},
		{"pass_with_slice", map[string]any{"items": []any{1, 2}}, map[string]any{"items": "required"}, false},
		{"pass_with_map", map[string]any{"meta": map[string]any{"k": "v"}}, map[string]any{"meta": "required"}, false},
		{"pass_zero_float", map[string]any{"val": 0.0}, map[string]any{"val": "required"}, false},
		{"fail_empty_string", map[string]any{"name": ""}, map[string]any{"name": "required"}, true},
		{"fail_whitespace_only", map[string]any{"name": "   "}, map[string]any{"name": "required"}, true},
		{"fail_nil", map[string]any{"name": nil}, map[string]any{"name": "required"}, true},
		{"fail_missing_key", map[string]any{"other": "x"}, map[string]any{"name": "required"}, true},
		{"pass_multiple_required", map[string]any{"a": "1", "b": "2"}, map[string]any{"a": "required", "b": "required"}, false},
		{"fail_one_of_multiple_missing", map[string]any{"a": "1"}, map[string]any{"a": "required", "b": "required"}, true},
		{"fail_empty_slice", map[string]any{"items": []any{}}, map[string]any{"items": "required"}, true},
		{"fail_empty_map", map[string]any{"data": map[string]any{}}, map[string]any{"data": "required"}, true},
		{"fail_nested_empty_string", map[string]any{"user": map[string]any{"name": ""}}, map[string]any{"user.name": "required"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_when_condition_met_and_present", map[string]any{"type": "admin", "role": "manager"}, map[string]any{"role": "required_if:type,admin"}, false},
		{"pass_when_condition_not_met", map[string]any{"type": "user"}, map[string]any{"role": "required_if:type,admin"}, false},
		{"fail_when_condition_met_and_missing", map[string]any{"type": "admin"}, map[string]any{"role": "required_if:type,admin"}, true},
		{"fail_when_condition_met_and_empty", map[string]any{"type": "admin", "role": ""}, map[string]any{"role": "required_if:type,admin"}, true},
		{"pass_multiple_values_no_match", map[string]any{"type": "guest"}, map[string]any{"role": "required_if:type,admin,manager"}, false},
		{"fail_multiple_values_match_second", map[string]any{"type": "manager", "role": ""}, map[string]any{"role": "required_if:type,admin,manager"}, true},
		{"pass_bool_condition_true", map[string]any{"active": true, "name": "go"}, map[string]any{"name": "required_if:active,true"}, false},
		{"fail_bool_condition_true_missing", map[string]any{"active": true}, map[string]any{"name": "required_if:active,true"}, true},
		{"pass_bool_condition_false_no_trigger", map[string]any{"active": false}, map[string]any{"name": "required_if:active,true"}, false},
		{"pass_other_field_missing", map[string]any{}, map[string]any{"role": "required_if:type,admin"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredUnless() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_when_other_matches", map[string]any{"role": "admin"}, map[string]any{"name": "required_unless:role,admin"}, false},
		{"pass_when_other_no_match_and_present", map[string]any{"role": "user", "name": "John"}, map[string]any{"name": "required_unless:role,admin"}, false},
		{"fail_when_other_no_match_and_missing", map[string]any{"role": "user"}, map[string]any{"name": "required_unless:role,admin"}, true},
		{"fail_when_other_no_match_and_empty", map[string]any{"role": "user", "name": ""}, map[string]any{"name": "required_unless:role,admin"}, true},
		{"pass_multiple_unless_match", map[string]any{"role": "mod"}, map[string]any{"name": "required_unless:role,admin,mod"}, false},
		{"fail_multiple_unless_no_match", map[string]any{"role": "user"}, map[string]any{"name": "required_unless:role,admin,mod"}, true},
		{"pass_other_field_missing_and_filled", map[string]any{"name": "go"}, map[string]any{"name": "required_unless:role,admin"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_when_other_present_and_filled", map[string]any{"first": "A", "last": "B"}, map[string]any{"last": "required_with:first"}, false},
		{"pass_when_other_absent", map[string]any{"other": "x"}, map[string]any{"last": "required_with:first"}, false},
		{"fail_when_other_present_and_empty", map[string]any{"first": "A", "last": ""}, map[string]any{"last": "required_with:first"}, true},
		{"fail_when_other_present_and_missing", map[string]any{"first": "A"}, map[string]any{"last": "required_with:first"}, true},
		{"pass_multiple_one_present", map[string]any{"a": "1", "c": "3"}, map[string]any{"c": "required_with:a,b"}, false},
		{"fail_multiple_one_present_target_missing", map[string]any{"a": "1"}, map[string]any{"c": "required_with:a,b"}, true},
		{"pass_other_present_but_empty_value", map[string]any{"first": "", "last": "B"}, map[string]any{"last": "required_with:first"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredWithAll() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_all_present_and_filled", map[string]any{"a": "1", "b": "2", "c": "3"}, map[string]any{"c": "required_with_all:a,b"}, false},
		{"pass_not_all_present", map[string]any{"a": "1"}, map[string]any{"c": "required_with_all:a,b"}, false},
		{"fail_all_present_and_empty", map[string]any{"a": "1", "b": "2", "c": ""}, map[string]any{"c": "required_with_all:a,b"}, true},
		{"fail_all_present_and_missing", map[string]any{"a": "1", "b": "2"}, map[string]any{"c": "required_with_all:a,b"}, true},
		{"pass_one_other_empty_str", map[string]any{"a": "1", "b": "", "c": "3"}, map[string]any{"c": "required_with_all:a,b"}, false},
		{"pass_three_others_not_all", map[string]any{"a": "1", "b": "2"}, map[string]any{"d": "required_with_all:a,b,c"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredWithout() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_when_other_absent_and_filled", map[string]any{"email": "a@b.com"}, map[string]any{"email": "required_without:phone"}, false},
		{"pass_when_other_present", map[string]any{"phone": "123"}, map[string]any{"email": "required_without:phone"}, false},
		{"fail_when_other_absent_and_empty", map[string]any{"other": "x"}, map[string]any{"email": "required_without:phone"}, true},
		{"fail_when_other_absent_and_missing", map[string]any{}, map[string]any{"email": "required_without:phone"}, true},
		{"pass_multiple_one_absent", map[string]any{"a": "1", "c": "3"}, map[string]any{"c": "required_without:a,b"}, false},
		{"pass_both_present", map[string]any{"a": "1", "b": "2", "c": "3"}, map[string]any{"c": "required_without:a,b"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredWithoutAll() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_none_present_and_filled", map[string]any{"c": "val"}, map[string]any{"c": "required_without_all:a,b"}, false},
		{"pass_some_present", map[string]any{"a": "1"}, map[string]any{"c": "required_without_all:a,b"}, false},
		{"pass_all_present", map[string]any{"a": "1", "b": "2"}, map[string]any{"c": "required_without_all:a,b"}, false},
		{"pass_one_of_three_present", map[string]any{"b": "2"}, map[string]any{"d": "required_without_all:a,b,c"}, false},
		{"fail_none_present_and_empty", map[string]any{"other": "x"}, map[string]any{"c": "required_without_all:a,b"}, true},
		{"fail_none_present_and_missing", map[string]any{}, map[string]any{"c": "required_without_all:a,b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredIfAccepted() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_accepted_and_filled", map[string]any{"terms": true, "sig": "yes"}, map[string]any{"sig": "required_if_accepted:terms"}, false},
		{"pass_not_accepted_bool", map[string]any{"terms": false}, map[string]any{"sig": "required_if_accepted:terms"}, false},
		{"pass_not_accepted_string", map[string]any{"terms": "no"}, map[string]any{"sig": "required_if_accepted:terms"}, false},
		{"pass_other_field_missing", map[string]any{}, map[string]any{"sig": "required_if_accepted:terms"}, false},
		{"fail_accepted_bool_and_missing", map[string]any{"terms": true}, map[string]any{"sig": "required_if_accepted:terms"}, true},
		{"fail_accepted_string_yes", map[string]any{"terms": "yes"}, map[string]any{"sig": "required_if_accepted:terms"}, true},
		{"fail_accepted_string_on", map[string]any{"terms": "on"}, map[string]any{"sig": "required_if_accepted:terms"}, true},
		{"fail_accepted_string_1", map[string]any{"terms": "1"}, map[string]any{"sig": "required_if_accepted:terms"}, true},
		{"fail_accepted_int_1", map[string]any{"terms": 1}, map[string]any{"sig": "required_if_accepted:terms"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRequiredIfDeclined() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_declined_and_filled", map[string]any{"auto": false, "reason": "manual"}, map[string]any{"reason": "required_if_declined:auto"}, false},
		{"pass_not_declined_bool", map[string]any{"auto": true}, map[string]any{"reason": "required_if_declined:auto"}, false},
		{"pass_not_declined_string", map[string]any{"auto": "yes"}, map[string]any{"reason": "required_if_declined:auto"}, false},
		{"pass_other_field_missing", map[string]any{}, map[string]any{"reason": "required_if_declined:auto"}, false},
		{"fail_declined_bool_and_missing", map[string]any{"auto": false}, map[string]any{"reason": "required_if_declined:auto"}, true},
		{"fail_declined_string_no", map[string]any{"auto": "no"}, map[string]any{"reason": "required_if_declined:auto"}, true},
		{"fail_declined_string_off", map[string]any{"auto": "off"}, map[string]any{"reason": "required_if_declined:auto"}, true},
		{"fail_declined_string_0", map[string]any{"auto": "0"}, map[string]any{"reason": "required_if_declined:auto"}, true},
		{"fail_declined_int_0", map[string]any{"auto": 0}, map[string]any{"reason": "required_if_declined:auto"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestFilled() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_present_with_value", map[string]any{"name": "go"}, map[string]any{"name": "filled"}, false},
		{"pass_present_with_int", map[string]any{"age": 0}, map[string]any{"age": "filled"}, false},
		{"pass_present_with_bool", map[string]any{"ok": false}, map[string]any{"ok": "filled"}, false},
		{"pass_not_present", map[string]any{"other": "x"}, map[string]any{"name": "filled"}, false},
		{"fail_present_empty", map[string]any{"name": ""}, map[string]any{"name": "filled"}, true},
		{"fail_present_whitespace", map[string]any{"name": "  "}, map[string]any{"name": "filled"}, true},
		{"fail_present_nil", map[string]any{"name": nil}, map[string]any{"name": "filled"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestPresent() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_present_with_value", map[string]any{"name": "go"}, map[string]any{"name": "present"}, false},
		{"pass_present_empty", map[string]any{"name": ""}, map[string]any{"name": "present"}, false},
		{"pass_present_nil", map[string]any{"name": nil}, map[string]any{"name": "present"}, false},
		{"fail_missing", map[string]any{"other": "x"}, map[string]any{"name": "present"}, true},
		{"fail_empty_data", map[string]any{}, map[string]any{"name": "present"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestPresentIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met_and_present", map[string]any{"type": "a", "name": ""}, map[string]any{"name": "present_if:type,a"}, false},
		{"pass_condition_not_met", map[string]any{"type": "b"}, map[string]any{"name": "present_if:type,a"}, false},
		{"fail_condition_met_and_missing", map[string]any{"type": "a"}, map[string]any{"name": "present_if:type,a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestPresentUnless() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met", map[string]any{"role": "admin"}, map[string]any{"name": "present_unless:role,admin"}, false},
		{"pass_condition_not_met_and_present", map[string]any{"role": "user", "name": ""}, map[string]any{"name": "present_unless:role,admin"}, false},
		{"fail_condition_not_met_and_missing", map[string]any{"role": "user"}, map[string]any{"name": "present_unless:role,admin"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestPresentWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_other_present_and_self_present", map[string]any{"a": "1", "b": ""}, map[string]any{"b": "present_with:a"}, false},
		{"pass_other_absent", map[string]any{"c": "1"}, map[string]any{"b": "present_with:a"}, false},
		{"fail_other_present_and_self_missing", map[string]any{"a": "1"}, map[string]any{"b": "present_with:a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestPresentWithAll() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_all_present_and_self_present", map[string]any{"a": "1", "b": "2", "c": ""}, map[string]any{"c": "present_with_all:a,b"}, false},
		{"pass_not_all_present", map[string]any{"a": "1"}, map[string]any{"c": "present_with_all:a,b"}, false},
		{"fail_all_present_and_self_missing", map[string]any{"a": "1", "b": "2"}, map[string]any{"c": "present_with_all:a,b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMissing() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_when_missing", map[string]any{"other": "x"}, map[string]any{"name": "missing"}, false},
		{"pass_empty_data", map[string]any{}, map[string]any{"name": "missing"}, false},
		{"fail_when_present", map[string]any{"name": "go"}, map[string]any{"name": "missing"}, true},
		{"fail_when_present_empty", map[string]any{"name": ""}, map[string]any{"name": "missing"}, true},
		{"fail_when_present_nil", map[string]any{"name": nil}, map[string]any{"name": "missing"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMissingIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met_and_missing", map[string]any{"type": "a"}, map[string]any{"name": "missing_if:type,a"}, false},
		{"pass_condition_not_met", map[string]any{"type": "b", "name": "go"}, map[string]any{"name": "missing_if:type,a"}, false},
		{"fail_condition_met_and_present", map[string]any{"type": "a", "name": "go"}, map[string]any{"name": "missing_if:type,a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMissingUnless() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met", map[string]any{"role": "admin", "name": "go"}, map[string]any{"name": "missing_unless:role,admin"}, false},
		{"pass_condition_not_met_and_missing", map[string]any{"role": "user"}, map[string]any{"name": "missing_unless:role,admin"}, false},
		{"fail_condition_not_met_and_present", map[string]any{"role": "user", "name": "go"}, map[string]any{"name": "missing_unless:role,admin"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMissingWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_other_present_and_self_missing", map[string]any{"a": "1"}, map[string]any{"b": "missing_with:a"}, false},
		{"pass_other_absent", map[string]any{"b": "1"}, map[string]any{"b": "missing_with:a"}, false},
		{"fail_other_present_and_self_present", map[string]any{"a": "1", "b": "2"}, map[string]any{"b": "missing_with:a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMissingWithAll() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_all_present_and_self_missing", map[string]any{"a": "1", "b": "2"}, map[string]any{"c": "missing_with_all:a,b"}, false},
		{"pass_not_all_present", map[string]any{"a": "1", "c": "3"}, map[string]any{"c": "missing_with_all:a,b"}, false},
		{"fail_all_present_and_self_present", map[string]any{"a": "1", "b": "2", "c": "3"}, map[string]any{"c": "missing_with_all:a,b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 2. Accept/Decline Rules =====

func (s *RulesTestSuite) TestAccepted() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_bool_true", map[string]any{"terms": true}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_yes", map[string]any{"terms": "yes"}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_on", map[string]any{"terms": "on"}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_1", map[string]any{"terms": "1"}, map[string]any{"terms": "accepted"}, false},
		{"pass_int_1", map[string]any{"terms": 1}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_true", map[string]any{"terms": "true"}, map[string]any{"terms": "accepted"}, false},
		{"pass_float_1", map[string]any{"terms": float64(1)}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_YES_case", map[string]any{"terms": "YES"}, map[string]any{"terms": "accepted"}, false},
		{"pass_string_True_case", map[string]any{"terms": "True"}, map[string]any{"terms": "accepted"}, false},
		{"fail_bool_false", map[string]any{"terms": false}, map[string]any{"terms": "accepted"}, true},
		{"fail_string_no", map[string]any{"terms": "no"}, map[string]any{"terms": "accepted"}, true},
		{"fail_int_0", map[string]any{"terms": 0}, map[string]any{"terms": "accepted"}, true},
		{"fail_nil", map[string]any{"terms": nil}, map[string]any{"terms": "accepted"}, true},
		{"fail_string_random", map[string]any{"terms": "maybe"}, map[string]any{"terms": "accepted"}, true},
		{"fail_int_2", map[string]any{"terms": 2}, map[string]any{"terms": "accepted"}, true},
		{"fail_empty_string", map[string]any{"terms": ""}, map[string]any{"terms": "accepted"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAcceptedIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met_and_accepted", map[string]any{"type": "a", "terms": true}, map[string]any{"terms": "accepted_if:type,a"}, false},
		{"pass_condition_not_met", map[string]any{"type": "b", "terms": false}, map[string]any{"terms": "accepted_if:type,a"}, false},
		{"fail_condition_met_and_not_accepted", map[string]any{"type": "a", "terms": false}, map[string]any{"terms": "accepted_if:type,a"}, true},
		{"fail_condition_met_and_missing", map[string]any{"type": "a"}, map[string]any{"terms": "accepted_if:type,a"}, true},
		{"pass_condition_met_string_yes", map[string]any{"type": "a", "terms": "yes"}, map[string]any{"terms": "accepted_if:type,a"}, false},
		{"pass_multiple_condition_values", map[string]any{"type": "b", "terms": "on"}, map[string]any{"terms": "accepted_if:type,a,b"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDeclined() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_bool_false", map[string]any{"opt": false}, map[string]any{"opt": "declined"}, false},
		{"pass_string_no", map[string]any{"opt": "no"}, map[string]any{"opt": "declined"}, false},
		{"pass_string_off", map[string]any{"opt": "off"}, map[string]any{"opt": "declined"}, false},
		{"pass_string_0", map[string]any{"opt": "0"}, map[string]any{"opt": "declined"}, false},
		{"pass_int_0", map[string]any{"opt": 0}, map[string]any{"opt": "declined"}, false},
		{"pass_string_false", map[string]any{"opt": "false"}, map[string]any{"opt": "declined"}, false},
		{"pass_float_0", map[string]any{"opt": float64(0)}, map[string]any{"opt": "declined"}, false},
		{"pass_string_NO_case", map[string]any{"opt": "NO"}, map[string]any{"opt": "declined"}, false},
		{"pass_string_False_case", map[string]any{"opt": "False"}, map[string]any{"opt": "declined"}, false},
		{"fail_bool_true", map[string]any{"opt": true}, map[string]any{"opt": "declined"}, true},
		{"fail_string_yes", map[string]any{"opt": "yes"}, map[string]any{"opt": "declined"}, true},
		{"fail_nil", map[string]any{"opt": nil}, map[string]any{"opt": "declined"}, true},
		{"fail_string_random", map[string]any{"opt": "maybe"}, map[string]any{"opt": "declined"}, true},
		{"fail_int_1", map[string]any{"opt": 1}, map[string]any{"opt": "declined"}, true},
		{"fail_empty_string", map[string]any{"opt": ""}, map[string]any{"opt": "declined"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDeclinedIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met_and_declined", map[string]any{"type": "a", "opt": false}, map[string]any{"opt": "declined_if:type,a"}, false},
		{"pass_condition_not_met", map[string]any{"type": "b", "opt": true}, map[string]any{"opt": "declined_if:type,a"}, false},
		{"fail_condition_met_and_not_declined", map[string]any{"type": "a", "opt": true}, map[string]any{"opt": "declined_if:type,a"}, true},
		{"fail_condition_met_and_missing", map[string]any{"type": "a"}, map[string]any{"opt": "declined_if:type,a"}, true},
		{"pass_condition_met_string_no", map[string]any{"type": "a", "opt": "no"}, map[string]any{"opt": "declined_if:type,a"}, false},
		{"pass_multiple_values", map[string]any{"type": "b", "opt": "off"}, map[string]any{"opt": "declined_if:type,a,b"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 3. Prohibition Rules =====

func (s *RulesTestSuite) TestProhibited() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_empty", map[string]any{"x": ""}, map[string]any{"x": "prohibited"}, false},
		{"pass_nil", map[string]any{"x": nil}, map[string]any{"x": "prohibited"}, false},
		{"pass_empty_slice", map[string]any{"x": []any{}}, map[string]any{"x": "prohibited"}, false},
		{"pass_empty_map", map[string]any{"x": map[string]any{}}, map[string]any{"x": "prohibited"}, false},
		{"fail_with_value", map[string]any{"x": "hello"}, map[string]any{"x": "prohibited"}, true},
		{"fail_with_int", map[string]any{"x": 1}, map[string]any{"x": "prohibited"}, true},
		{"fail_with_bool", map[string]any{"x": true}, map[string]any{"x": "prohibited"}, true},
		{"fail_with_slice", map[string]any{"x": []any{1}}, map[string]any{"x": "prohibited"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestProhibitedIf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met_and_empty", map[string]any{"type": "a", "x": ""}, map[string]any{"x": "prohibited_if:type,a"}, false},
		{"pass_condition_not_met", map[string]any{"type": "b", "x": "val"}, map[string]any{"x": "prohibited_if:type,a"}, false},
		{"fail_condition_met_and_has_value", map[string]any{"type": "a", "x": "val"}, map[string]any{"x": "prohibited_if:type,a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestProhibitedUnless() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_condition_met", map[string]any{"role": "admin", "x": "val"}, map[string]any{"x": "prohibited_unless:role,admin"}, false},
		{"pass_condition_not_met_and_empty", map[string]any{"role": "user", "x": ""}, map[string]any{"x": "prohibited_unless:role,admin"}, false},
		{"fail_condition_not_met_and_has_value", map[string]any{"role": "user", "x": "val"}, map[string]any{"x": "prohibited_unless:role,admin"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestProhibitedIfAccepted() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_accepted_and_empty", map[string]any{"terms": true, "x": ""}, map[string]any{"x": "prohibited_if_accepted:terms"}, false},
		{"pass_not_accepted", map[string]any{"terms": false, "x": "val"}, map[string]any{"x": "prohibited_if_accepted:terms"}, false},
		{"fail_accepted_and_has_value", map[string]any{"terms": true, "x": "val"}, map[string]any{"x": "prohibited_if_accepted:terms"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestProhibitedIfDeclined() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_declined_and_empty", map[string]any{"auto": false, "x": ""}, map[string]any{"x": "prohibited_if_declined:auto"}, false},
		{"pass_not_declined", map[string]any{"auto": true, "x": "val"}, map[string]any{"x": "prohibited_if_declined:auto"}, false},
		{"fail_declined_and_has_value", map[string]any{"auto": false, "x": "val"}, map[string]any{"x": "prohibited_if_declined:auto"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestProhibits() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_self_empty", map[string]any{"x": "", "y": "val"}, map[string]any{"x": "prohibits:y"}, false},
		{"pass_self_has_value_other_empty", map[string]any{"x": "val", "y": ""}, map[string]any{"x": "prohibits:y"}, false},
		{"pass_self_has_value_other_missing", map[string]any{"x": "val"}, map[string]any{"x": "prohibits:y"}, false},
		{"pass_self_nil", map[string]any{"x": nil, "y": "val"}, map[string]any{"x": "prohibits:y"}, false},
		{"fail_both_have_values", map[string]any{"x": "val", "y": "val2"}, map[string]any{"x": "prohibits:y"}, true},
		{"fail_multiple_one_present", map[string]any{"x": "val", "y": "", "z": "val2"}, map[string]any{"x": "prohibits:y,z"}, true},
		{"pass_multiple_all_empty", map[string]any{"x": "val", "y": "", "z": ""}, map[string]any{"x": "prohibits:y,z"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 4. Type Rules =====

func (s *RulesTestSuite) TestStringRule() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_string", map[string]any{"x": "hello"}, map[string]any{"x": "string"}, false},
		{"pass_string_with_spaces", map[string]any{"x": "hello world"}, map[string]any{"x": "string"}, false},
		{"pass_string_unicode", map[string]any{"x": "你好世界"}, map[string]any{"x": "string"}, false},
		{"pass_string_numeric_str", map[string]any{"x": "12345"}, map[string]any{"x": "string"}, false},
		{"fail_int", map[string]any{"x": 123}, map[string]any{"x": "string"}, true},
		{"fail_bool", map[string]any{"x": true}, map[string]any{"x": "string"}, true},
		{"fail_float", map[string]any{"x": 3.14}, map[string]any{"x": "string"}, true},
		{"fail_slice", map[string]any{"x": []any{1}}, map[string]any{"x": "string"}, true},
		{"fail_map", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "string"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestInteger() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_int", map[string]any{"x": 42}, map[string]any{"x": "integer"}, false},
		{"pass_int64", map[string]any{"x": int64(42)}, map[string]any{"x": "integer"}, false},
		{"pass_int32", map[string]any{"x": int32(42)}, map[string]any{"x": "integer"}, false},
		{"pass_int8", map[string]any{"x": int8(42)}, map[string]any{"x": "integer"}, false},
		{"pass_uint", map[string]any{"x": uint(42)}, map[string]any{"x": "integer"}, false},
		{"pass_string_int", map[string]any{"x": "42"}, map[string]any{"x": "integer"}, false},
		{"pass_string_negative", map[string]any{"x": "-10"}, map[string]any{"x": "integer"}, false},
		{"pass_float_whole", map[string]any{"x": 42.0}, map[string]any{"x": "integer"}, false},
		{"pass_zero", map[string]any{"x": 0}, map[string]any{"x": "integer"}, false},
		{"fail_float_decimal", map[string]any{"x": 42.5}, map[string]any{"x": "integer"}, true},
		{"fail_string_alpha", map[string]any{"x": "abc"}, map[string]any{"x": "integer"}, true},
		{"fail_string_float", map[string]any{"x": "42.5"}, map[string]any{"x": "integer"}, true},
		{"fail_bool", map[string]any{"x": true}, map[string]any{"x": "integer"}, true},
		{"pass_int_alias", map[string]any{"x": 1}, map[string]any{"x": "int"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUintRule() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_uint", map[string]any{"v": uint(42)}, map[string]any{"v": "uint"}, false},
		{"pass_uint8", map[string]any{"v": uint8(255)}, map[string]any{"v": "uint"}, false},
		{"pass_uint16", map[string]any{"v": uint16(65535)}, map[string]any{"v": "uint"}, false},
		{"pass_uint32", map[string]any{"v": uint32(100)}, map[string]any{"v": "uint"}, false},
		{"pass_uint64", map[string]any{"v": uint64(100)}, map[string]any{"v": "uint"}, false},
		{"pass_int_positive", map[string]any{"v": 42}, map[string]any{"v": "uint"}, false},
		{"pass_int8_positive", map[string]any{"v": int8(1)}, map[string]any{"v": "uint"}, false},
		{"pass_int16_positive", map[string]any{"v": int16(1)}, map[string]any{"v": "uint"}, false},
		{"pass_int32_positive", map[string]any{"v": int32(1)}, map[string]any{"v": "uint"}, false},
		{"pass_int64_positive", map[string]any{"v": int64(1)}, map[string]any{"v": "uint"}, false},
		{"pass_zero", map[string]any{"v": 0}, map[string]any{"v": "uint"}, false},
		{"pass_string_zero", map[string]any{"v": "0"}, map[string]any{"v": "uint"}, false},
		{"pass_string", map[string]any{"v": "42"}, map[string]any{"v": "uint"}, false},
		{"pass_float64_whole", map[string]any{"v": float64(42)}, map[string]any{"v": "uint"}, false},
		{"fail_negative", map[string]any{"v": -1}, map[string]any{"v": "uint"}, true},
		{"fail_negative_int64", map[string]any{"v": int64(-1)}, map[string]any{"v": "uint"}, true},
		{"fail_float", map[string]any{"v": 3.14}, map[string]any{"v": "uint"}, true},
		{"fail_string_alpha", map[string]any{"v": "abc"}, map[string]any{"v": "uint"}, true},
		{"fail_negative_string", map[string]any{"v": "-5"}, map[string]any{"v": "uint"}, true},
		{"fail_float_string", map[string]any{"v": "3.14"}, map[string]any{"v": "uint"}, true},
		{"fail_bool", map[string]any{"v": true}, map[string]any{"v": "uint"}, true},
		{"fail_slice", map[string]any{"v": []any{1}}, map[string]any{"v": "uint"}, true},
		{"fail_map", map[string]any{"v": map[string]any{}}, map[string]any{"v": "uint"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestNumeric() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_int", map[string]any{"x": 42}, map[string]any{"x": "numeric"}, false},
		{"pass_float", map[string]any{"x": 3.14}, map[string]any{"x": "numeric"}, false},
		{"pass_numeric_string", map[string]any{"x": "3.14"}, map[string]any{"x": "numeric"}, false},
		{"pass_negative_string", map[string]any{"x": "-5"}, map[string]any{"x": "numeric"}, false},
		{"pass_zero", map[string]any{"x": 0}, map[string]any{"x": "numeric"}, false},
		{"pass_int64", map[string]any{"x": int64(100)}, map[string]any{"x": "numeric"}, false},
		{"pass_uint", map[string]any{"x": uint(50)}, map[string]any{"x": "numeric"}, false},
		{"pass_string_int", map[string]any{"x": "42"}, map[string]any{"x": "numeric"}, false},
		{"pass_bool_converts", map[string]any{"x": true}, map[string]any{"x": "numeric"}, false},
		{"fail_alpha_string", map[string]any{"x": "abc"}, map[string]any{"x": "numeric"}, true},
		{"fail_mixed_string", map[string]any{"x": "12abc"}, map[string]any{"x": "numeric"}, true},
		{"fail_slice", map[string]any{"x": []any{1}}, map[string]any{"x": "numeric"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestBooleanRule() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_true", map[string]any{"x": true}, map[string]any{"x": "boolean"}, false},
		{"pass_false", map[string]any{"x": false}, map[string]any{"x": "boolean"}, false},
		{"pass_int_0", map[string]any{"x": 0}, map[string]any{"x": "boolean"}, false},
		{"pass_int_1", map[string]any{"x": 1}, map[string]any{"x": "boolean"}, false},
		{"pass_int64_0", map[string]any{"x": int64(0)}, map[string]any{"x": "boolean"}, false},
		{"pass_int64_1", map[string]any{"x": int64(1)}, map[string]any{"x": "boolean"}, false},
		{"pass_float64_0", map[string]any{"x": float64(0)}, map[string]any{"x": "boolean"}, false},
		{"pass_float64_1", map[string]any{"x": float64(1)}, map[string]any{"x": "boolean"}, false},
		{"pass_string_true", map[string]any{"x": "true"}, map[string]any{"x": "boolean"}, false},
		{"pass_string_false", map[string]any{"x": "false"}, map[string]any{"x": "boolean"}, false},
		{"pass_string_0", map[string]any{"x": "0"}, map[string]any{"x": "boolean"}, false},
		{"pass_string_1", map[string]any{"x": "1"}, map[string]any{"x": "boolean"}, false},
		{"pass_string_yes", map[string]any{"x": "yes"}, map[string]any{"x": "boolean"}, false},
		{"pass_string_on", map[string]any{"x": "on"}, map[string]any{"x": "boolean"}, false},
		{"fail_int_2", map[string]any{"x": 2}, map[string]any{"x": "boolean"}, true},
		{"fail_int_neg1", map[string]any{"x": -1}, map[string]any{"x": "boolean"}, true},
		{"fail_string_2", map[string]any{"x": "2"}, map[string]any{"x": "boolean"}, true},
		{"fail_slice", map[string]any{"x": []any{true}}, map[string]any{"x": "boolean"}, true},
		{"pass_bool_alias", map[string]any{"x": true}, map[string]any{"x": "bool"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestFloat() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_float64", map[string]any{"x": 3.14}, map[string]any{"x": "float"}, false},
		{"pass_float32", map[string]any{"x": float32(3.14)}, map[string]any{"x": "float"}, false},
		{"pass_string_float", map[string]any{"x": "3.14"}, map[string]any{"x": "float"}, false},
		{"pass_string_negative_float", map[string]any{"x": "-3.14"}, map[string]any{"x": "float"}, false},
		{"pass_string_whole_number", map[string]any{"x": "42"}, map[string]any{"x": "float"}, false},
		{"pass_float64_zero", map[string]any{"x": 0.0}, map[string]any{"x": "float"}, false},
		{"fail_int", map[string]any{"x": 42}, map[string]any{"x": "float"}, true},
		{"fail_string", map[string]any{"x": "abc"}, map[string]any{"x": "float"}, true},
		{"fail_bool", map[string]any{"x": true}, map[string]any{"x": "float"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestArray() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_slice", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array"}, false},
		{"pass_map", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "array"}, false},
		{"pass_string_slice", map[string]any{"x": []string{"a"}}, map[string]any{"x": "array"}, false},
		{"pass_empty_slice", map[string]any{"x": []any{}}, map[string]any{"x": "array"}, false},
		{"pass_empty_map", map[string]any{"x": map[string]any{}}, map[string]any{"x": "array"}, false},
		{"pass_int_slice", map[string]any{"x": []int{1, 2}}, map[string]any{"x": "array"}, false},
		{"pass_float64_slice", map[string]any{"x": []float64{1.1, 2.2}}, map[string]any{"x": "array"}, false},
		{"pass_bool_slice", map[string]any{"x": []bool{true, false}}, map[string]any{"x": "array"}, false},
		{"pass_fixed_array", map[string]any{"x": [3]int{1, 2, 3}}, map[string]any{"x": "array"}, false},
		{"fail_string", map[string]any{"x": "hello"}, map[string]any{"x": "array"}, true},
		{"fail_int", map[string]any{"x": 42}, map[string]any{"x": "array"}, true},
		{"fail_nil", map[string]any{"x": nil}, map[string]any{"x": "array"}, true},
		{"fail_bool", map[string]any{"x": true}, map[string]any{"x": "array"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestList() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_slice", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "list"}, false},
		{"pass_empty_slice", map[string]any{"x": []any{}}, map[string]any{"x": "list"}, false},
		{"pass_string_slice", map[string]any{"x": []string{"a"}}, map[string]any{"x": "list"}, false},
		{"fail_map", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "list"}, true},
		{"fail_string", map[string]any{"x": "hello"}, map[string]any{"x": "list"}, true},
		{"fail_nil", map[string]any{"x": nil}, map[string]any{"x": "list"}, true},
		{"pass_slice_alias", map[string]any{"x": []any{1}}, map[string]any{"x": "slice"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMap() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_map", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "map"}, false},
		{"pass_empty_map", map[string]any{"x": map[string]any{}}, map[string]any{"x": "map"}, false},
		{"fail_slice", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "map"}, true},
		{"fail_string", map[string]any{"x": "hello"}, map[string]any{"x": "map"}, true},
		{"fail_nil", map[string]any{"x": nil}, map[string]any{"x": "map"}, true},
		{"fail_int", map[string]any{"x": 42}, map[string]any{"x": "map"}, true},
		{"fail_bool", map[string]any{"x": true}, map[string]any{"x": "map"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 5. Size Rules =====

func (s *RulesTestSuite) TestSize() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String type: size = character count
		{"pass_string_exact", map[string]any{"x": "abc"}, map[string]any{"x": "string|size:3"}, false},
		{"fail_string_too_short", map[string]any{"x": "ab"}, map[string]any{"x": "string|size:3"}, true},
		{"fail_string_too_long", map[string]any{"x": "abcd"}, map[string]any{"x": "string|size:3"}, true},
		{"pass_string_empty_size_zero", map[string]any{"x": ""}, map[string]any{"x": "string|size:0"}, false},
		{"pass_string_unicode", map[string]any{"x": "你好世"}, map[string]any{"x": "string|size:3"}, false},
		{"pass_string_single_char", map[string]any{"x": "a"}, map[string]any{"x": "string|size:1"}, false},

		// Numeric type: size = numeric value
		{"pass_numeric_exact_int", map[string]any{"x": 10}, map[string]any{"x": "numeric|size:10"}, false},
		{"fail_numeric_wrong", map[string]any{"x": 5}, map[string]any{"x": "numeric|size:10"}, true},
		{"pass_numeric_exact_float", map[string]any{"x": 3.14}, map[string]any{"x": "numeric|size:3.14"}, false},
		{"pass_numeric_zero", map[string]any{"x": 0}, map[string]any{"x": "numeric|size:0"}, false},
		{"pass_numeric_negative", map[string]any{"x": -5}, map[string]any{"x": "numeric|size:-5"}, false},
		{"fail_numeric_negative_mismatch", map[string]any{"x": -3}, map[string]any{"x": "numeric|size:-5"}, true},
		{"pass_numeric_string_value", map[string]any{"x": "42"}, map[string]any{"x": "numeric|size:42"}, false},
		{"pass_numeric_float64", map[string]any{"x": float64(100)}, map[string]any{"x": "numeric|size:100"}, false},

		// Array type: size = element count
		{"pass_array_exact", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|size:3"}, false},
		{"fail_array_too_few", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|size:3"}, true},
		{"fail_array_too_many", map[string]any{"x": []any{1, 2, 3, 4}}, map[string]any{"x": "array|size:3"}, true},
		{"pass_array_empty_size_zero", map[string]any{"x": []any{}}, map[string]any{"x": "array|size:0"}, false},
		{"pass_array_single", map[string]any{"x": []any{"a"}}, map[string]any{"x": "array|size:1"}, false},

		// Map type with size
		{"pass_map_exact", map[string]any{"x": map[string]any{"a": 1, "b": 2}}, map[string]any{"x": "map|size:2"}, false},
		{"fail_map_wrong", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "map|size:2"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMin() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String: min character count
		{"pass_string_above", map[string]any{"x": "abcde"}, map[string]any{"x": "string|min:3"}, false},
		{"pass_string_exact", map[string]any{"x": "abc"}, map[string]any{"x": "string|min:3"}, false},
		{"fail_string_below", map[string]any{"x": "ab"}, map[string]any{"x": "string|min:3"}, true},
		{"pass_string_unicode_above", map[string]any{"x": "你好世界"}, map[string]any{"x": "string|min:3"}, false},
		{"fail_string_unicode_below", map[string]any{"x": "你好"}, map[string]any{"x": "string|min:3"}, true},
		{"pass_string_min_zero", map[string]any{"x": ""}, map[string]any{"x": "string|min:0"}, false},
		{"pass_string_min_one", map[string]any{"x": "a"}, map[string]any{"x": "string|min:1"}, false},

		// Numeric: min numeric value
		{"pass_numeric_above", map[string]any{"x": 15}, map[string]any{"x": "numeric|min:10"}, false},
		{"pass_numeric_exact", map[string]any{"x": 10}, map[string]any{"x": "numeric|min:10"}, false},
		{"fail_numeric_below", map[string]any{"x": 5}, map[string]any{"x": "numeric|min:10"}, true},
		{"pass_numeric_float", map[string]any{"x": 3.5}, map[string]any{"x": "numeric|min:3.5"}, false},
		{"fail_numeric_float_below", map[string]any{"x": 3.4}, map[string]any{"x": "numeric|min:3.5"}, true},
		{"pass_numeric_negative", map[string]any{"x": -1}, map[string]any{"x": "numeric|min:-5"}, false},
		{"fail_numeric_negative_below", map[string]any{"x": -10}, map[string]any{"x": "numeric|min:-5"}, true},
		{"pass_numeric_zero", map[string]any{"x": 0}, map[string]any{"x": "numeric|min:0"}, false},
		{"pass_numeric_string", map[string]any{"x": "100"}, map[string]any{"x": "numeric|min:50"}, false},

		// Array: min element count
		{"pass_array_above", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|min:2"}, false},
		{"pass_array_exact", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|min:2"}, false},
		{"fail_array_below", map[string]any{"x": []any{1}}, map[string]any{"x": "array|min:2"}, true},
		{"pass_array_min_zero", map[string]any{"x": []any{}}, map[string]any{"x": "array|min:0"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMax() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String
		{"pass_string_below", map[string]any{"x": "ab"}, map[string]any{"x": "string|max:3"}, false},
		{"pass_string_exact", map[string]any{"x": "abc"}, map[string]any{"x": "string|max:3"}, false},
		{"fail_string_above", map[string]any{"x": "abcd"}, map[string]any{"x": "string|max:3"}, true},
		{"pass_string_empty", map[string]any{"x": ""}, map[string]any{"x": "string|max:3"}, false},
		{"pass_string_unicode_within", map[string]any{"x": "你好"}, map[string]any{"x": "string|max:3"}, false},
		{"fail_string_unicode_over", map[string]any{"x": "你好世界"}, map[string]any{"x": "string|max:3"}, true},

		// Numeric
		{"pass_numeric_below", map[string]any{"x": 5}, map[string]any{"x": "numeric|max:10"}, false},
		{"pass_numeric_exact", map[string]any{"x": 10}, map[string]any{"x": "numeric|max:10"}, false},
		{"fail_numeric_above", map[string]any{"x": 15}, map[string]any{"x": "numeric|max:10"}, true},
		{"pass_numeric_zero", map[string]any{"x": 0}, map[string]any{"x": "numeric|max:10"}, false},
		{"pass_numeric_negative", map[string]any{"x": -5}, map[string]any{"x": "numeric|max:0"}, false},
		{"fail_numeric_negative_threshold", map[string]any{"x": -3}, map[string]any{"x": "numeric|max:-5"}, true},
		{"pass_numeric_float", map[string]any{"x": 3.14}, map[string]any{"x": "numeric|max:3.14"}, false},
		{"fail_numeric_float_over", map[string]any{"x": 3.15}, map[string]any{"x": "numeric|max:3.14"}, true},

		// Array
		{"pass_array_below", map[string]any{"x": []any{1}}, map[string]any{"x": "array|max:2"}, false},
		{"pass_array_exact", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|max:2"}, false},
		{"fail_array_above", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|max:2"}, true},
		{"pass_array_empty", map[string]any{"x": []any{}}, map[string]any{"x": "array|max:2"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestNumericAliasesWithSizeRules() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// int alias should resolve size as numeric value, not string length
		{"int_max_pass", map[string]any{"x": 3}, map[string]any{"x": "int|max:3"}, false},
		{"int_max_fail", map[string]any{"x": 42}, map[string]any{"x": "int|max:3"}, true},
		{"int_max_string_input_fail", map[string]any{"x": "42"}, map[string]any{"x": "int|max:3"}, true},
		{"int_min_pass", map[string]any{"x": 10}, map[string]any{"x": "int|min:5"}, false},
		{"int_min_fail", map[string]any{"x": 2}, map[string]any{"x": "int|min:5"}, true},
		{"int_between_pass", map[string]any{"x": 5}, map[string]any{"x": "int|between:1,10"}, false},
		{"int_between_fail", map[string]any{"x": 20}, map[string]any{"x": "int|between:1,10"}, true},

		// uint alias
		{"uint_max_pass", map[string]any{"x": 3}, map[string]any{"x": "uint|max:5"}, false},
		{"uint_max_fail", map[string]any{"x": 10}, map[string]any{"x": "uint|max:5"}, true},
		{"uint_min_pass", map[string]any{"x": 5}, map[string]any{"x": "uint|min:0"}, false},

		// float alias
		{"float_max_pass", map[string]any{"x": 3.14}, map[string]any{"x": "float|max:5"}, false},
		{"float_max_fail", map[string]any{"x": 10.5}, map[string]any{"x": "float|max:5"}, true},
		{"float_min_pass", map[string]any{"x": 3.14}, map[string]any{"x": "float|min:3.14"}, false},
		{"float_min_fail", map[string]any{"x": 1.5}, map[string]any{"x": "float|min:3.14"}, true},
		{"float_size_pass", map[string]any{"x": 3.14}, map[string]any{"x": "float|size:3.14"}, false},
		{"float_size_fail", map[string]any{"x": 2.0}, map[string]any{"x": "float|size:3.14"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails(), tt.name)
		})
	}
}

func (s *RulesTestSuite) TestBetween() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String
		{"pass_string_in_range", map[string]any{"x": "abcd"}, map[string]any{"x": "string|between:3,5"}, false},
		{"pass_string_at_min", map[string]any{"x": "abc"}, map[string]any{"x": "string|between:3,5"}, false},
		{"pass_string_at_max", map[string]any{"x": "abcde"}, map[string]any{"x": "string|between:3,5"}, false},
		{"fail_string_below", map[string]any{"x": "ab"}, map[string]any{"x": "string|between:3,5"}, true},
		{"fail_string_above", map[string]any{"x": "abcdef"}, map[string]any{"x": "string|between:3,5"}, true},

		// Numeric
		{"pass_numeric_in_range", map[string]any{"x": 25}, map[string]any{"x": "numeric|between:18,65"}, false},
		{"pass_numeric_at_min", map[string]any{"x": 18}, map[string]any{"x": "numeric|between:18,65"}, false},
		{"pass_numeric_at_max", map[string]any{"x": 65}, map[string]any{"x": "numeric|between:18,65"}, false},
		{"fail_numeric_below", map[string]any{"x": 10}, map[string]any{"x": "numeric|between:18,65"}, true},
		{"fail_numeric_above", map[string]any{"x": 100}, map[string]any{"x": "numeric|between:18,65"}, true},
		{"pass_numeric_float_range", map[string]any{"x": 1.5}, map[string]any{"x": "numeric|between:1.0,2.0"}, false},
		{"pass_numeric_negative_range", map[string]any{"x": -3}, map[string]any{"x": "numeric|between:-5,-1"}, false},

		// Array
		{"pass_array_in_range", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|between:2,4"}, false},
		{"pass_array_at_min", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|between:2,4"}, false},
		{"pass_array_at_max", map[string]any{"x": []any{1, 2, 3, 4}}, map[string]any{"x": "array|between:2,4"}, false},
		{"fail_array_below", map[string]any{"x": []any{1}}, map[string]any{"x": "array|between:2,4"}, true},
		{"fail_array_above", map[string]any{"x": []any{1, 2, 3, 4, 5}}, map[string]any{"x": "array|between:2,4"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestGt() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String: character count > threshold
		{"pass_string_greater", map[string]any{"x": "abcde"}, map[string]any{"x": "string|gt:3"}, false},
		{"fail_string_equal", map[string]any{"x": "abc"}, map[string]any{"x": "string|gt:3"}, true},
		{"fail_string_less", map[string]any{"x": "ab"}, map[string]any{"x": "string|gt:3"}, true},

		// Numeric: value > threshold
		{"pass_numeric_greater", map[string]any{"x": 15}, map[string]any{"x": "numeric|gt:10"}, false},
		{"fail_numeric_equal", map[string]any{"x": 10}, map[string]any{"x": "numeric|gt:10"}, true},
		{"fail_numeric_less", map[string]any{"x": 5}, map[string]any{"x": "numeric|gt:10"}, true},
		{"pass_numeric_float_greater", map[string]any{"x": 3.15}, map[string]any{"x": "numeric|gt:3.14"}, false},
		{"fail_numeric_float_equal", map[string]any{"x": 3.14}, map[string]any{"x": "numeric|gt:3.14"}, true},

		// Array: count > threshold
		{"pass_array_greater", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|gt:2"}, false},
		{"fail_array_equal", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|gt:2"}, true},

		// Field reference: compare against another field
		{"pass_gt_field_ref_string", map[string]any{"x": "abcd", "y": "ab"}, map[string]any{"x": "string|gt:y", "y": "string"}, false},
		{"fail_gt_field_ref_string", map[string]any{"x": "ab", "y": "abcd"}, map[string]any{"x": "string|gt:y", "y": "string"}, true},
		{"pass_gt_field_ref_numeric", map[string]any{"x": 20, "y": 10}, map[string]any{"x": "numeric|gt:y", "y": "numeric"}, false},
		{"fail_gt_field_ref_numeric", map[string]any{"x": 10, "y": 10}, map[string]any{"x": "numeric|gt:y", "y": "numeric"}, true},
		{"pass_gt_field_ref_array", map[string]any{"x": []any{1, 2, 3}, "y": []any{1}}, map[string]any{"x": "array|gt:y", "y": "array"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestGte() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String
		{"pass_string_greater", map[string]any{"x": "abcde"}, map[string]any{"x": "string|gte:3"}, false},
		{"pass_string_equal", map[string]any{"x": "abc"}, map[string]any{"x": "string|gte:3"}, false},
		{"fail_string_less", map[string]any{"x": "ab"}, map[string]any{"x": "string|gte:3"}, true},

		// Numeric
		{"pass_numeric_greater", map[string]any{"x": 15}, map[string]any{"x": "numeric|gte:10"}, false},
		{"pass_numeric_equal", map[string]any{"x": 10}, map[string]any{"x": "numeric|gte:10"}, false},
		{"fail_numeric_less", map[string]any{"x": 5}, map[string]any{"x": "numeric|gte:10"}, true},
		{"pass_numeric_float_equal", map[string]any{"x": 3.14}, map[string]any{"x": "numeric|gte:3.14"}, false},
		{"pass_numeric_negative", map[string]any{"x": -3}, map[string]any{"x": "numeric|gte:-5"}, false},

		// Array
		{"pass_array_greater", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|gte:2"}, false},
		{"pass_array_equal", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|gte:2"}, false},
		{"fail_array_less", map[string]any{"x": []any{1}}, map[string]any{"x": "array|gte:2"}, true},

		// Field reference
		{"pass_gte_field_ref_numeric", map[string]any{"x": 10, "y": 10}, map[string]any{"x": "numeric|gte:y", "y": "numeric"}, false},
		{"pass_gte_field_ref_numeric_greater", map[string]any{"x": 15, "y": 10}, map[string]any{"x": "numeric|gte:y", "y": "numeric"}, false},
		{"fail_gte_field_ref_numeric", map[string]any{"x": 5, "y": 10}, map[string]any{"x": "numeric|gte:y", "y": "numeric"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestLt() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String
		{"pass_string_less", map[string]any{"x": "ab"}, map[string]any{"x": "string|lt:3"}, false},
		{"fail_string_equal", map[string]any{"x": "abc"}, map[string]any{"x": "string|lt:3"}, true},
		{"fail_string_greater", map[string]any{"x": "abcd"}, map[string]any{"x": "string|lt:3"}, true},

		// Numeric
		{"pass_numeric_less", map[string]any{"x": 5}, map[string]any{"x": "numeric|lt:10"}, false},
		{"fail_numeric_equal", map[string]any{"x": 10}, map[string]any{"x": "numeric|lt:10"}, true},
		{"fail_numeric_greater", map[string]any{"x": 15}, map[string]any{"x": "numeric|lt:10"}, true},
		{"pass_numeric_negative", map[string]any{"x": -10}, map[string]any{"x": "numeric|lt:-5"}, false},
		{"fail_numeric_negative_equal", map[string]any{"x": -5}, map[string]any{"x": "numeric|lt:-5"}, true},

		// Array
		{"pass_array_less", map[string]any{"x": []any{1}}, map[string]any{"x": "array|lt:2"}, false},
		{"fail_array_equal", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|lt:2"}, true},

		// Field reference
		{"pass_lt_field_ref_numeric", map[string]any{"x": 5, "y": 10}, map[string]any{"x": "numeric|lt:y", "y": "numeric"}, false},
		{"fail_lt_field_ref_numeric", map[string]any{"x": 10, "y": 10}, map[string]any{"x": "numeric|lt:y", "y": "numeric"}, true},
		{"pass_lt_field_ref_string", map[string]any{"x": "ab", "y": "abcde"}, map[string]any{"x": "string|lt:y", "y": "string"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestLte() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// String
		{"pass_string_less", map[string]any{"x": "ab"}, map[string]any{"x": "string|lte:3"}, false},
		{"pass_string_equal", map[string]any{"x": "abc"}, map[string]any{"x": "string|lte:3"}, false},
		{"fail_string_greater", map[string]any{"x": "abcd"}, map[string]any{"x": "string|lte:3"}, true},

		// Numeric
		{"pass_numeric_less", map[string]any{"x": 5}, map[string]any{"x": "numeric|lte:10"}, false},
		{"pass_numeric_equal", map[string]any{"x": 10}, map[string]any{"x": "numeric|lte:10"}, false},
		{"fail_numeric_greater", map[string]any{"x": 15}, map[string]any{"x": "numeric|lte:10"}, true},
		{"pass_numeric_zero", map[string]any{"x": 0}, map[string]any{"x": "numeric|lte:0"}, false},
		{"pass_numeric_negative", map[string]any{"x": -10}, map[string]any{"x": "numeric|lte:-5"}, false},
		{"pass_numeric_negative_equal", map[string]any{"x": -5}, map[string]any{"x": "numeric|lte:-5"}, false},

		// Array
		{"pass_array_less", map[string]any{"x": []any{1}}, map[string]any{"x": "array|lte:2"}, false},
		{"pass_array_equal", map[string]any{"x": []any{1, 2}}, map[string]any{"x": "array|lte:2"}, false},
		{"fail_array_greater", map[string]any{"x": []any{1, 2, 3}}, map[string]any{"x": "array|lte:2"}, true},

		// Field reference
		{"pass_lte_field_ref_numeric_equal", map[string]any{"x": 10, "y": 10}, map[string]any{"x": "numeric|lte:y", "y": "numeric"}, false},
		{"pass_lte_field_ref_numeric_less", map[string]any{"x": 5, "y": 10}, map[string]any{"x": "numeric|lte:y", "y": "numeric"}, false},
		{"fail_lte_field_ref_numeric", map[string]any{"x": 15, "y": 10}, map[string]any{"x": "numeric|lte:y", "y": "numeric"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 6. Numeric Rules =====

func (s *RulesTestSuite) TestDigits() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_exact_digits_3", map[string]any{"x": "123"}, map[string]any{"x": "digits:3"}, false},
		{"pass_exact_digits_1", map[string]any{"x": "5"}, map[string]any{"x": "digits:1"}, false},
		{"pass_exact_digits_6", map[string]any{"x": "000000"}, map[string]any{"x": "digits:6"}, false},
		{"fail_wrong_count_less", map[string]any{"x": "12"}, map[string]any{"x": "digits:3"}, true},
		{"fail_wrong_count_more", map[string]any{"x": "1234"}, map[string]any{"x": "digits:3"}, true},
		{"fail_non_digit_letter", map[string]any{"x": "12a"}, map[string]any{"x": "digits:3"}, true},
		{"fail_non_digit_symbol", map[string]any{"x": "12."}, map[string]any{"x": "digits:3"}, true},
		{"fail_negative_sign", map[string]any{"x": "-12"}, map[string]any{"x": "digits:3"}, true},
		{"pass_int_value", map[string]any{"x": 1234}, map[string]any{"x": "digits:4"}, false},
		{"pass_int_single_digit", map[string]any{"x": 0}, map[string]any{"x": "digits:1"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDigitsBetween() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_in_range", map[string]any{"x": "123"}, map[string]any{"x": "digits_between:2,4"}, false},
		{"pass_at_min", map[string]any{"x": "12"}, map[string]any{"x": "digits_between:2,4"}, false},
		{"pass_at_max", map[string]any{"x": "1234"}, map[string]any{"x": "digits_between:2,4"}, false},
		{"fail_below_min", map[string]any{"x": "1"}, map[string]any{"x": "digits_between:2,4"}, true},
		{"fail_above_max", map[string]any{"x": "12345"}, map[string]any{"x": "digits_between:2,4"}, true},
		{"fail_non_digit", map[string]any{"x": "12a"}, map[string]any{"x": "digits_between:2,4"}, true},
		{"pass_int_value", map[string]any{"x": 123}, map[string]any{"x": "digits_between:2,4"}, false},
		{"pass_all_zeros", map[string]any{"x": "000"}, map[string]any{"x": "digits_between:2,4"}, false},
		{"pass_range_1_10", map[string]any{"x": "12345"}, map[string]any{"x": "digits_between:1,10"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDecimal() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_exact_2_places", map[string]any{"x": "3.14"}, map[string]any{"x": "decimal:2"}, false},
		{"fail_1_place_for_2", map[string]any{"x": "3.1"}, map[string]any{"x": "decimal:2"}, true},
		{"fail_3_places_for_2", map[string]any{"x": "3.142"}, map[string]any{"x": "decimal:2"}, true},
		{"pass_range_1_3", map[string]any{"x": "3.14"}, map[string]any{"x": "decimal:1,3"}, false},
		{"pass_range_min", map[string]any{"x": "3.1"}, map[string]any{"x": "decimal:1,3"}, false},
		{"pass_range_max", map[string]any{"x": "3.142"}, map[string]any{"x": "decimal:1,3"}, false},
		{"fail_range_below", map[string]any{"x": "3"}, map[string]any{"x": "decimal:1,3"}, true},
		{"fail_range_above", map[string]any{"x": "3.1415"}, map[string]any{"x": "decimal:1,3"}, true},
		{"pass_zero_decimal_places", map[string]any{"x": "3"}, map[string]any{"x": "decimal:0"}, false},
		{"fail_has_decimal_for_0", map[string]any{"x": "3.1"}, map[string]any{"x": "decimal:0"}, true},
		{"fail_not_numeric", map[string]any{"x": "abc"}, map[string]any{"x": "decimal:2"}, true},
		{"pass_negative_decimal", map[string]any{"x": "-3.14"}, map[string]any{"x": "decimal:2"}, false},
		{"pass_zero_value", map[string]any{"x": "0.00"}, map[string]any{"x": "decimal:2"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMultipleOf() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_15_of_5", map[string]any{"x": 15}, map[string]any{"x": "multiple_of:5"}, false},
		{"pass_10_of_5", map[string]any{"x": 10}, map[string]any{"x": "multiple_of:5"}, false},
		{"fail_13_of_5", map[string]any{"x": 13}, map[string]any{"x": "multiple_of:5"}, true},
		{"pass_zero_of_5", map[string]any{"x": 0}, map[string]any{"x": "multiple_of:5"}, false},
		{"pass_100_of_25", map[string]any{"x": 100}, map[string]any{"x": "multiple_of:25"}, false},
		{"fail_99_of_25", map[string]any{"x": 99}, map[string]any{"x": "multiple_of:25"}, true},
		{"pass_negative_of_3", map[string]any{"x": -9}, map[string]any{"x": "multiple_of:3"}, false},
		{"fail_negative_of_4", map[string]any{"x": -9}, map[string]any{"x": "multiple_of:4"}, true},
		{"pass_float_of_half", map[string]any{"x": 1.5}, map[string]any{"x": "multiple_of:0.5"}, false},
		{"pass_string_number", map[string]any{"x": "12"}, map[string]any{"x": "multiple_of:4"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMinDigits() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_exact_digits", map[string]any{"x": "123"}, map[string]any{"x": "min_digits:3"}, false},
		{"pass_more_digits", map[string]any{"x": "12345"}, map[string]any{"x": "min_digits:3"}, false},
		{"fail_too_few", map[string]any{"x": "12"}, map[string]any{"x": "min_digits:3"}, true},
		{"pass_int_value", map[string]any{"x": 12345}, map[string]any{"x": "min_digits:3"}, false},
		{"fail_single_digit", map[string]any{"x": "5"}, map[string]any{"x": "min_digits:3"}, true},
		{"pass_min_1", map[string]any{"x": "0"}, map[string]any{"x": "min_digits:1"}, false},
		{"pass_with_leading_zeros", map[string]any{"x": "001"}, map[string]any{"x": "min_digits:3"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMaxDigits() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_within_limit", map[string]any{"x": "12"}, map[string]any{"x": "max_digits:3"}, false},
		{"pass_at_limit", map[string]any{"x": "123"}, map[string]any{"x": "max_digits:3"}, false},
		{"fail_over_limit", map[string]any{"x": "1234"}, map[string]any{"x": "max_digits:3"}, true},
		{"pass_single_digit", map[string]any{"x": "5"}, map[string]any{"x": "max_digits:3"}, false},
		{"pass_int_value", map[string]any{"x": 99}, map[string]any{"x": "max_digits:3"}, false},
		{"fail_int_over", map[string]any{"x": 10000}, map[string]any{"x": "max_digits:3"}, true},
		{"pass_max_1", map[string]any{"x": "0"}, map[string]any{"x": "max_digits:1"}, false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 7. String Format Rules =====

func (s *RulesTestSuite) TestAlpha() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_letters", map[string]any{"x": "Hello"}, map[string]any{"x": "alpha"}, false},
		{"pass_unicode_letters", map[string]any{"x": "Héllo"}, map[string]any{"x": "alpha"}, false},
		{"pass_chinese", map[string]any{"x": "你好"}, map[string]any{"x": "alpha"}, false},
		{"pass_single_letter", map[string]any{"x": "a"}, map[string]any{"x": "alpha"}, false},
		{"fail_with_numbers", map[string]any{"x": "abc123"}, map[string]any{"x": "alpha"}, true},
		{"fail_with_spaces", map[string]any{"x": "abc def"}, map[string]any{"x": "alpha"}, true},
		{"fail_with_special", map[string]any{"x": "abc!"}, map[string]any{"x": "alpha"}, true},
		{"fail_with_dash", map[string]any{"x": "abc-def"}, map[string]any{"x": "alpha"}, true},
		{"fail_with_underscore", map[string]any{"x": "abc_def"}, map[string]any{"x": "alpha"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAlphaNum() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_letters_and_numbers", map[string]any{"x": "abc123"}, map[string]any{"x": "alpha_num"}, false},
		{"pass_letters_only", map[string]any{"x": "abc"}, map[string]any{"x": "alpha_num"}, false},
		{"pass_numbers_only", map[string]any{"x": "123"}, map[string]any{"x": "alpha_num"}, false},
		{"pass_unicode_letters_nums", map[string]any{"x": "Héllo123"}, map[string]any{"x": "alpha_num"}, false},
		{"fail_with_spaces", map[string]any{"x": "abc 123"}, map[string]any{"x": "alpha_num"}, true},
		{"fail_with_special", map[string]any{"x": "abc@123"}, map[string]any{"x": "alpha_num"}, true},
		{"fail_with_dash", map[string]any{"x": "abc-123"}, map[string]any{"x": "alpha_num"}, true},
		{"fail_with_underscore", map[string]any{"x": "abc_123"}, map[string]any{"x": "alpha_num"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAlphaDash() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_with_dash_underscore", map[string]any{"x": "abc-123_def"}, map[string]any{"x": "alpha_dash"}, false},
		{"pass_letters_only", map[string]any{"x": "abc"}, map[string]any{"x": "alpha_dash"}, false},
		{"pass_numbers_only", map[string]any{"x": "123"}, map[string]any{"x": "alpha_dash"}, false},
		{"pass_underscore_only", map[string]any{"x": "___"}, map[string]any{"x": "alpha_dash"}, false},
		{"pass_dash_only", map[string]any{"x": "---"}, map[string]any{"x": "alpha_dash"}, false},
		{"pass_slug", map[string]any{"x": "my-blog-post"}, map[string]any{"x": "alpha_dash"}, false},
		{"fail_with_spaces", map[string]any{"x": "abc def"}, map[string]any{"x": "alpha_dash"}, true},
		{"fail_with_special", map[string]any{"x": "abc@def"}, map[string]any{"x": "alpha_dash"}, true},
		{"fail_with_dot", map[string]any{"x": "abc.def"}, map[string]any{"x": "alpha_dash"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAscii() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_ascii_text", map[string]any{"x": "hello world"}, map[string]any{"x": "ascii"}, false},
		{"pass_ascii_with_symbols", map[string]any{"x": "hello@world.com!#$%"}, map[string]any{"x": "ascii"}, false},
		{"pass_ascii_numbers", map[string]any{"x": "12345"}, map[string]any{"x": "ascii"}, false},
		{"pass_ascii_newline_tab", map[string]any{"x": "hello\nworld\t!"}, map[string]any{"x": "ascii"}, false},
		{"fail_unicode_accent", map[string]any{"x": "héllo"}, map[string]any{"x": "ascii"}, true},
		{"fail_emoji", map[string]any{"x": "hello 🎉"}, map[string]any{"x": "ascii"}, true},
		{"fail_chinese", map[string]any{"x": "你好"}, map[string]any{"x": "ascii"}, true},
		{"fail_japanese", map[string]any{"x": "こんにちは"}, map[string]any{"x": "ascii"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestEmail() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_simple", map[string]any{"x": "user@example.com"}, map[string]any{"x": "email"}, false},
		{"pass_with_plus", map[string]any{"x": "user+tag@example.com"}, map[string]any{"x": "email"}, false},
		{"pass_with_dot", map[string]any{"x": "first.last@example.com"}, map[string]any{"x": "email"}, false},
		{"pass_subdomain", map[string]any{"x": "user@mail.example.com"}, map[string]any{"x": "email"}, false},
		{"pass_numbers", map[string]any{"x": "user123@example.com"}, map[string]any{"x": "email"}, false},
		{"pass_dash_domain", map[string]any{"x": "user@my-domain.com"}, map[string]any{"x": "email"}, false},
		{"fail_no_at", map[string]any{"x": "userexample.com"}, map[string]any{"x": "email"}, true},
		{"fail_no_domain", map[string]any{"x": "user@"}, map[string]any{"x": "email"}, true},
		{"fail_no_user", map[string]any{"x": "@example.com"}, map[string]any{"x": "email"}, true},
		{"fail_plain_string", map[string]any{"x": "notanemail"}, map[string]any{"x": "email"}, true},
		{"fail_double_at", map[string]any{"x": "user@@example.com"}, map[string]any{"x": "email"}, true},
		{"fail_spaces", map[string]any{"x": "user @example.com"}, map[string]any{"x": "email"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUrl() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_https", map[string]any{"x": "https://goravel.dev"}, map[string]any{"x": "url"}, false},
		{"pass_http", map[string]any{"x": "https://example.com/path"}, map[string]any{"x": "url"}, false},
		{"pass_with_port", map[string]any{"x": "https://localhost:8080"}, map[string]any{"x": "url"}, false},
		{"pass_with_query", map[string]any{"x": "https://example.com/path?q=test&a=1"}, map[string]any{"x": "url"}, false},
		{"pass_with_fragment", map[string]any{"x": "https://example.com/page#section"}, map[string]any{"x": "url"}, false},
		{"pass_ftp", map[string]any{"x": "ftp://files.example.com"}, map[string]any{"x": "url"}, false},
		{"fail_no_scheme", map[string]any{"x": "goravel.dev"}, map[string]any{"x": "url"}, true},
		{"fail_plain_string", map[string]any{"x": "not a url"}, map[string]any{"x": "url"}, true},
		{"fail_just_path", map[string]any{"x": "/path/to/file"}, map[string]any{"x": "url"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestIp() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_ipv4", map[string]any{"x": "192.168.1.1"}, map[string]any{"x": "ip"}, false},
		{"pass_ipv4_loopback", map[string]any{"x": "127.0.0.1"}, map[string]any{"x": "ip"}, false},
		{"pass_ipv4_all_zeros", map[string]any{"x": "0.0.0.0"}, map[string]any{"x": "ip"}, false},
		{"pass_ipv6_full", map[string]any{"x": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}, map[string]any{"x": "ip"}, false},
		{"pass_ipv6_loopback", map[string]any{"x": "::1"}, map[string]any{"x": "ip"}, false},
		{"fail_invalid", map[string]any{"x": "not-an-ip"}, map[string]any{"x": "ip"}, true},
		{"fail_out_of_range", map[string]any{"x": "999.999.999.999"}, map[string]any{"x": "ip"}, true},
		{"fail_incomplete", map[string]any{"x": "192.168.1"}, map[string]any{"x": "ip"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestIpv4() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_valid", map[string]any{"x": "192.168.1.1"}, map[string]any{"x": "ipv4"}, false},
		{"pass_loopback", map[string]any{"x": "127.0.0.1"}, map[string]any{"x": "ipv4"}, false},
		{"pass_broadcast", map[string]any{"x": "255.255.255.255"}, map[string]any{"x": "ipv4"}, false},
		{"pass_all_zeros", map[string]any{"x": "0.0.0.0"}, map[string]any{"x": "ipv4"}, false},
		{"fail_ipv6", map[string]any{"x": "::1"}, map[string]any{"x": "ipv4"}, true},
		{"fail_invalid", map[string]any{"x": "abc"}, map[string]any{"x": "ipv4"}, true},
		{"fail_three_octets", map[string]any{"x": "192.168.1"}, map[string]any{"x": "ipv4"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestIpv6() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_full", map[string]any{"x": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}, map[string]any{"x": "ipv6"}, false},
		{"pass_loopback", map[string]any{"x": "::1"}, map[string]any{"x": "ipv6"}, false},
		{"pass_abbreviated", map[string]any{"x": "fe80::1"}, map[string]any{"x": "ipv6"}, false},
		{"pass_all_zeros", map[string]any{"x": "::"}, map[string]any{"x": "ipv6"}, false},
		{"fail_ipv4", map[string]any{"x": "192.168.1.1"}, map[string]any{"x": "ipv6"}, true},
		{"fail_invalid", map[string]any{"x": "not-ipv6"}, map[string]any{"x": "ipv6"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMacAddress() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_colon", map[string]any{"x": "00:1B:44:11:3A:B7"}, map[string]any{"x": "mac_address"}, false},
		{"pass_dash", map[string]any{"x": "00-1B-44-11-3A-B7"}, map[string]any{"x": "mac_address"}, false},
		{"pass_lowercase_mac", map[string]any{"x": "00:1b:44:11:3a:b7"}, map[string]any{"x": "mac_address"}, false},
		{"pass_alias_mac", map[string]any{"x": "00:1B:44:11:3A:B7"}, map[string]any{"x": "mac"}, false},
		{"fail_invalid", map[string]any{"x": "not-a-mac"}, map[string]any{"x": "mac_address"}, true},
		{"fail_too_short", map[string]any{"x": "00:1B:44"}, map[string]any{"x": "mac_address"}, true},
		{"fail_invalid_hex", map[string]any{"x": "GG:HH:II:JJ:KK:LL"}, map[string]any{"x": "mac_address"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestJson() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_object", map[string]any{"x": `{"key":"val"}`}, map[string]any{"x": "json"}, false},
		{"pass_array", map[string]any{"x": `[1,2,3]`}, map[string]any{"x": "json"}, false},
		{"pass_string", map[string]any{"x": `"hello"`}, map[string]any{"x": "json"}, false},
		{"pass_number", map[string]any{"x": `42`}, map[string]any{"x": "json"}, false},
		{"pass_boolean", map[string]any{"x": `true`}, map[string]any{"x": "json"}, false},
		{"pass_null", map[string]any{"x": `null`}, map[string]any{"x": "json"}, false},
		{"pass_nested", map[string]any{"x": `{"a":{"b":[1,2]}}`}, map[string]any{"x": "json"}, false},
		{"pass_empty_object", map[string]any{"x": `{}`}, map[string]any{"x": "json"}, false},
		{"pass_empty_array", map[string]any{"x": `[]`}, map[string]any{"x": "json"}, false},
		{"fail_invalid_braces", map[string]any{"x": `{invalid}`}, map[string]any{"x": "json"}, true},
		{"fail_plain", map[string]any{"x": "hello"}, map[string]any{"x": "json"}, true},
		{"fail_trailing_comma", map[string]any{"x": `{"a":1,}`}, map[string]any{"x": "json"}, true},
		{"fail_single_quotes", map[string]any{"x": `{'a': 1}`}, map[string]any{"x": "json"}, true},
		{"fail_non_string_type", map[string]any{"x": 123}, map[string]any{"x": "json"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUuid() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_v4", map[string]any{"x": "550e8400-e29b-41d4-a716-446655440000"}, map[string]any{"x": "uuid"}, false},
		{"pass_v3", map[string]any{"x": "a3bb189e-8bf9-3888-9912-ace4e6543002"}, map[string]any{"x": "uuid"}, false},
		{"pass_v5", map[string]any{"x": "886313e1-3b8a-5372-9b90-0c9aee199e5d"}, map[string]any{"x": "uuid"}, false},
		{"pass_v1", map[string]any{"x": "550e8400-e29b-11d4-a716-446655440000"}, map[string]any{"x": "uuid"}, false},
		{"pass_lowercase", map[string]any{"x": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}, map[string]any{"x": "uuid"}, false},
		{"pass_uppercase", map[string]any{"x": "A0EEBC99-9C0B-4EF8-BB6D-6BB9BD380A11"}, map[string]any{"x": "uuid"}, false},
		{"pass_nil_uuid", map[string]any{"x": "00000000-0000-0000-0000-000000000000"}, map[string]any{"x": "uuid"}, false},
		{"fail_invalid", map[string]any{"x": "not-a-uuid"}, map[string]any{"x": "uuid"}, true},
		{"fail_missing_dashes", map[string]any{"x": "550e8400e29b41d4a716446655440000"}, map[string]any{"x": "uuid"}, true},
		{"fail_too_short", map[string]any{"x": "550e8400-e29b"}, map[string]any{"x": "uuid"}, true},
		{"fail_too_long", map[string]any{"x": "550e8400-e29b-41d4-a716-446655440000a"}, map[string]any{"x": "uuid"}, true},
		{"fail_invalid_chars", map[string]any{"x": "gggggggg-gggg-gggg-gggg-gggggggggggg"}, map[string]any{"x": "uuid"}, true},
		{"fail_int", map[string]any{"x": 123}, map[string]any{"x": "uuid"}, true},
		{"fail_nil", map[string]any{"x": nil}, map[string]any{"x": "uuid"}, true},
		{"fail_extra_dash", map[string]any{"x": "550e8400-e29b-41d4-a716-4466554400-00"}, map[string]any{"x": "uuid"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUuid3() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass", map[string]any{"v": "a3bb189e-8bf9-3888-9912-ace4e6543002"}, map[string]any{"v": "uuid3"}, false},
		{"pass_uppercase", map[string]any{"v": "A3BB189E-8BF9-3888-9912-ACE4E6543002"}, map[string]any{"v": "uuid3"}, false},
		{"fail_uuid4", map[string]any{"v": "550e8400-e29b-41d4-a716-446655440000"}, map[string]any{"v": "uuid3"}, true},
		{"fail_uuid5", map[string]any{"v": "886313e1-3b8a-5372-9b90-0c9aee199e5d"}, map[string]any{"v": "uuid3"}, true},
		{"fail_uuid1", map[string]any{"v": "550e8400-e29b-11d4-a716-446655440000"}, map[string]any{"v": "uuid3"}, true},
		{"fail_not_uuid", map[string]any{"v": "not-a-uuid"}, map[string]any{"v": "uuid3"}, true},
		{"fail_int", map[string]any{"v": 123}, map[string]any{"v": "uuid3"}, true},
		{"fail_nil", map[string]any{"v": nil}, map[string]any{"v": "uuid3"}, true},
		{"fail_too_short", map[string]any{"v": "a3bb189e-8bf9-3888"}, map[string]any{"v": "uuid3"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUuid4() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass", map[string]any{"v": "550e8400-e29b-41d4-a716-446655440000"}, map[string]any{"v": "uuid4"}, false},
		{"pass_variant_8", map[string]any{"v": "550e8400-e29b-41d4-8716-446655440000"}, map[string]any{"v": "uuid4"}, false},
		{"pass_variant_9", map[string]any{"v": "550e8400-e29b-41d4-9716-446655440000"}, map[string]any{"v": "uuid4"}, false},
		{"pass_variant_b", map[string]any{"v": "550e8400-e29b-41d4-b716-446655440000"}, map[string]any{"v": "uuid4"}, false},
		{"pass_uppercase", map[string]any{"v": "550E8400-E29B-41D4-A716-446655440000"}, map[string]any{"v": "uuid4"}, false},
		{"fail_uuid3", map[string]any{"v": "a3bb189e-8bf9-3888-9912-ace4e6543002"}, map[string]any{"v": "uuid4"}, true},
		{"fail_uuid5", map[string]any{"v": "886313e1-3b8a-5372-9b90-0c9aee199e5d"}, map[string]any{"v": "uuid4"}, true},
		{"fail_uuid1", map[string]any{"v": "550e8400-e29b-11d4-a716-446655440000"}, map[string]any{"v": "uuid4"}, true},
		{"fail_bad_variant", map[string]any{"v": "550e8400-e29b-41d4-0716-446655440000"}, map[string]any{"v": "uuid4"}, true},
		{"fail_not_uuid", map[string]any{"v": "not-a-uuid"}, map[string]any{"v": "uuid4"}, true},
		{"fail_int", map[string]any{"v": 123}, map[string]any{"v": "uuid4"}, true},
		{"fail_nil", map[string]any{"v": nil}, map[string]any{"v": "uuid4"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUuid5() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass", map[string]any{"v": "886313e1-3b8a-5372-9b90-0c9aee199e5d"}, map[string]any{"v": "uuid5"}, false},
		{"pass_variant_8", map[string]any{"v": "886313e1-3b8a-5372-8b90-0c9aee199e5d"}, map[string]any{"v": "uuid5"}, false},
		{"pass_variant_b", map[string]any{"v": "886313e1-3b8a-5372-bb90-0c9aee199e5d"}, map[string]any{"v": "uuid5"}, false},
		{"pass_uppercase", map[string]any{"v": "886313E1-3B8A-5372-9B90-0C9AEE199E5D"}, map[string]any{"v": "uuid5"}, false},
		{"fail_uuid3", map[string]any{"v": "a3bb189e-8bf9-3888-9912-ace4e6543002"}, map[string]any{"v": "uuid5"}, true},
		{"fail_uuid4", map[string]any{"v": "550e8400-e29b-41d4-a716-446655440000"}, map[string]any{"v": "uuid5"}, true},
		{"fail_uuid1", map[string]any{"v": "550e8400-e29b-11d4-a716-446655440000"}, map[string]any{"v": "uuid5"}, true},
		{"fail_bad_variant", map[string]any{"v": "886313e1-3b8a-5372-0b90-0c9aee199e5d"}, map[string]any{"v": "uuid5"}, true},
		{"fail_not_uuid", map[string]any{"v": "not-a-uuid"}, map[string]any{"v": "uuid5"}, true},
		{"fail_int", map[string]any{"v": 123}, map[string]any{"v": "uuid5"}, true},
		{"fail_nil", map[string]any{"v": nil}, map[string]any{"v": "uuid5"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUlid() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_valid", map[string]any{"x": "01ARZ3NDEKTSV4RRFFQ69G5FAV"}, map[string]any{"x": "ulid"}, false},
		{"pass_lowercase_ulid", map[string]any{"x": "01arz3ndektsv4rrffq69g5fav"}, map[string]any{"x": "ulid"}, false},
		{"fail_invalid", map[string]any{"x": "not-a-ulid"}, map[string]any{"x": "ulid"}, true},
		{"fail_too_short", map[string]any{"x": "01ARZ3NDEK"}, map[string]any{"x": "ulid"}, true},
		{"fail_too_long", map[string]any{"x": "01ARZ3NDEKTSV4RRFFQ69G5FAVX"}, map[string]any{"x": "ulid"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestHexColor() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_6_digit", map[string]any{"x": "#FF5733"}, map[string]any{"x": "hex_color"}, false},
		{"pass_3_digit", map[string]any{"x": "#FFF"}, map[string]any{"x": "hex_color"}, false},
		{"pass_8_digit_alpha", map[string]any{"x": "#FF5733AA"}, map[string]any{"x": "hex_color"}, false},
		{"pass_lowercase_hex", map[string]any{"x": "#ff5733"}, map[string]any{"x": "hex_color"}, false},
		{"fail_no_hash", map[string]any{"x": "FF5733"}, map[string]any{"x": "hex_color"}, true},
		{"fail_invalid_chars", map[string]any{"x": "#GGGGGG"}, map[string]any{"x": "hex_color"}, true},
		{"fail_empty_hash", map[string]any{"x": "#"}, map[string]any{"x": "hex_color"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestRegex() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_match", map[string]any{"x": "abc123"}, map[string]any{"x": `regex:^[a-z]+\d+$`}, false},
		{"pass_simple_digits", map[string]any{"x": "12345"}, map[string]any{"x": `regex:^\d+$`}, false},
		{"pass_email_pattern", map[string]any{"x": "a@b.c"}, map[string]any{"x": `regex:^.+@.+\..+$`}, false},
		{"fail_no_match", map[string]any{"x": "ABC"}, map[string]any{"x": `regex:^[a-z]+\d+$`}, true},
		{"fail_partial_match", map[string]any{"x": "abc"}, map[string]any{"x": `regex:^[a-z]+\d+$`}, true},
		// Array syntax for regex with pipe
		{"pass_regex_with_pipe_array", map[string]any{"x": "foo"}, map[string]any{"x": []string{"required", `regex:^(foo|bar)$`}}, false},
		{"fail_regex_with_pipe_array", map[string]any{"x": "baz"}, map[string]any{"x": []string{"required", `regex:^(foo|bar)$`}}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestNotRegex() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_no_match", map[string]any{"x": "ABC"}, map[string]any{"x": `not_regex:^[a-z]+$`}, false},
		{"pass_numbers", map[string]any{"x": "123"}, map[string]any{"x": `not_regex:^[a-z]+$`}, false},
		{"fail_match", map[string]any{"x": "abc"}, map[string]any{"x": `not_regex:^[a-z]+$`}, true},
		{"fail_full_match", map[string]any{"x": "hello"}, map[string]any{"x": `not_regex:^hello$`}, true},
		// Array syntax for not_regex with pipe
		{"pass_not_regex_pipe_array", map[string]any{"x": "baz"}, map[string]any{"x": []string{"required", `not_regex:^(foo|bar)$`}}, false},
		{"fail_not_regex_pipe_array", map[string]any{"x": "foo"}, map[string]any{"x": []string{"required", `not_regex:^(foo|bar)$`}}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestLowercase() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_lowercase", map[string]any{"x": "hello"}, map[string]any{"x": "lowercase"}, false},
		{"pass_lowercase_with_numbers", map[string]any{"x": "hello123"}, map[string]any{"x": "lowercase"}, false},
		{"pass_lowercase_with_spaces", map[string]any{"x": "hello world"}, map[string]any{"x": "lowercase"}, false},
		{"pass_lowercase_with_symbols", map[string]any{"x": "hello@world!"}, map[string]any{"x": "lowercase"}, false},
		{"fail_uppercase_first", map[string]any{"x": "Hello"}, map[string]any{"x": "lowercase"}, true},
		{"fail_all_uppercase", map[string]any{"x": "HELLO"}, map[string]any{"x": "lowercase"}, true},
		{"fail_mixed", map[string]any{"x": "hELLO"}, map[string]any{"x": "lowercase"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestUppercase() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_uppercase", map[string]any{"x": "HELLO"}, map[string]any{"x": "uppercase"}, false},
		{"pass_uppercase_with_numbers", map[string]any{"x": "HELLO123"}, map[string]any{"x": "uppercase"}, false},
		{"pass_uppercase_with_spaces", map[string]any{"x": "HELLO WORLD"}, map[string]any{"x": "uppercase"}, false},
		{"pass_uppercase_with_symbols", map[string]any{"x": "HELLO@WORLD!"}, map[string]any{"x": "uppercase"}, false},
		{"fail_lowercase_first", map[string]any{"x": "hELLO"}, map[string]any{"x": "uppercase"}, true},
		{"fail_all_lowercase", map[string]any{"x": "hello"}, map[string]any{"x": "uppercase"}, true},
		{"fail_mixed", map[string]any{"x": "Hello"}, map[string]any{"x": "uppercase"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 8. String Content Rules =====

func (s *RulesTestSuite) TestStartsWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_single", map[string]any{"x": "hello world"}, map[string]any{"x": "starts_with:hello"}, false},
		{"pass_one_of", map[string]any{"x": "world hello"}, map[string]any{"x": "starts_with:hello,world"}, false},
		{"pass_exact", map[string]any{"x": "hello"}, map[string]any{"x": "starts_with:hello"}, false},
		{"fail_empty_prefix_no_params", map[string]any{"x": "anything"}, map[string]any{"x": "starts_with:"}, true},
		{"fail_none", map[string]any{"x": "goodbye"}, map[string]any{"x": "starts_with:hello,world"}, true},
		{"fail_case_sensitive", map[string]any{"x": "Hello"}, map[string]any{"x": "starts_with:hello"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDoesntStartWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_no_match", map[string]any{"x": "goodbye"}, map[string]any{"x": "doesnt_start_with:hello,world"}, false},
		{"pass_different_case", map[string]any{"x": "Hello"}, map[string]any{"x": "doesnt_start_with:hello"}, false},
		{"fail_match_first", map[string]any{"x": "hello world"}, map[string]any{"x": "doesnt_start_with:hello,world"}, true},
		{"fail_match_second", map[string]any{"x": "world hello"}, map[string]any{"x": "doesnt_start_with:hello,world"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestEndsWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_single", map[string]any{"x": "hello world"}, map[string]any{"x": "ends_with:world"}, false},
		{"pass_one_of", map[string]any{"x": "test.jpg"}, map[string]any{"x": "ends_with:.jpg,.png"}, false},
		{"pass_exact", map[string]any{"x": "world"}, map[string]any{"x": "ends_with:world"}, false},
		{"fail_none", map[string]any{"x": "test.gif"}, map[string]any{"x": "ends_with:.jpg,.png"}, true},
		{"fail_case_sensitive", map[string]any{"x": "test.JPG"}, map[string]any{"x": "ends_with:.jpg"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDoesntEndWith() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_no_match", map[string]any{"x": "test.gif"}, map[string]any{"x": "doesnt_end_with:.jpg,.png"}, false},
		{"pass_different_case", map[string]any{"x": "test.JPG"}, map[string]any{"x": "doesnt_end_with:.jpg"}, false},
		{"fail_match_first", map[string]any{"x": "test.jpg"}, map[string]any{"x": "doesnt_end_with:.jpg,.png"}, true},
		{"fail_match_second", map[string]any{"x": "test.png"}, map[string]any{"x": "doesnt_end_with:.jpg,.png"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestContains() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_contains", map[string]any{"x": "hello world"}, map[string]any{"x": "contains:world"}, false},
		{"pass_contains_all", map[string]any{"x": "hello world foo"}, map[string]any{"x": "contains:hello,world"}, false},
		{"pass_contains_single", map[string]any{"x": "abcdef"}, map[string]any{"x": "contains:cd"}, false},
		{"fail_missing_one", map[string]any{"x": "hello foo"}, map[string]any{"x": "contains:hello,world"}, true},
		{"fail_missing_all", map[string]any{"x": "goodbye"}, map[string]any{"x": "contains:hello,world"}, true},
		{"fail_case_sensitive", map[string]any{"x": "Hello"}, map[string]any{"x": "contains:hello"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDoesntContain() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_no_match", map[string]any{"x": "hello world"}, map[string]any{"x": "doesnt_contain:foo,bar"}, false},
		{"pass_case_different", map[string]any{"x": "Hello World"}, map[string]any{"x": "doesnt_contain:hello"}, false},
		{"fail_match_one", map[string]any{"x": "hello world"}, map[string]any{"x": "doesnt_contain:hello,bar"}, true},
		{"fail_match_multiple", map[string]any{"x": "hello world"}, map[string]any{"x": "doesnt_contain:hello,world"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestConfirmed() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_matching", map[string]any{"pw": "secret", "pw_confirmation": "secret"}, map[string]any{"pw": "confirmed"}, false},
		{"pass_matching_number", map[string]any{"code": 123, "code_confirmation": 123}, map[string]any{"code": "confirmed"}, false},
		{"fail_mismatch", map[string]any{"pw": "secret", "pw_confirmation": "diff"}, map[string]any{"pw": "confirmed"}, true},
		{"fail_missing_confirmation", map[string]any{"pw": "secret"}, map[string]any{"pw": "confirmed"}, true},
		{"fail_empty_confirmation", map[string]any{"pw": "secret", "pw_confirmation": ""}, map[string]any{"pw": "confirmed"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 9. Comparison Rules =====

func (s *RulesTestSuite) TestSame() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_same_string", map[string]any{"a": "x", "b": "x"}, map[string]any{"a": "same:b"}, false},
		{"pass_same_int", map[string]any{"a": 42, "b": 42}, map[string]any{"a": "same:b"}, false},
		{"pass_same_bool", map[string]any{"a": true, "b": true}, map[string]any{"a": "same:b"}, false},
		{"pass_string_int_same_via_sprintf", map[string]any{"a": "1", "b": 1}, map[string]any{"a": "same:b"}, false},
		{"fail_different_value", map[string]any{"a": "x", "b": "y"}, map[string]any{"a": "same:b"}, true},
		{"fail_missing_other", map[string]any{"a": "x"}, map[string]any{"a": "same:b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDifferent() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_different_value", map[string]any{"a": "x", "b": "y"}, map[string]any{"a": "different:b"}, false},
		{"pass_missing_other", map[string]any{"a": "x"}, map[string]any{"a": "different:b"}, false},
		{"fail_string_int_same_via_sprintf", map[string]any{"a": "1", "b": 1}, map[string]any{"a": "different:b"}, true},
		{"fail_same_string", map[string]any{"a": "x", "b": "x"}, map[string]any{"a": "different:b"}, true},
		{"fail_same_int", map[string]any{"a": 42, "b": 42}, map[string]any{"a": "different:b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestEq() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_string", map[string]any{"v": "hello"}, map[string]any{"v": "eq:hello"}, false},
		{"pass_int", map[string]any{"v": 42}, map[string]any{"v": "eq:42"}, false},
		{"pass_float", map[string]any{"v": 3.14}, map[string]any{"v": "eq:3.14"}, false},
		{"pass_bool_true", map[string]any{"v": true}, map[string]any{"v": "eq:true"}, false},
		{"pass_bool_false", map[string]any{"v": false}, map[string]any{"v": "eq:false"}, false},
		{"pass_zero", map[string]any{"v": 0}, map[string]any{"v": "eq:0"}, false},
		{"pass_empty_string", map[string]any{"v": ""}, map[string]any{"v": "eq:"}, false},
		{"fail_different_string", map[string]any{"v": "hello"}, map[string]any{"v": "eq:world"}, true},
		{"fail_different_int", map[string]any{"v": 42}, map[string]any{"v": "eq:43"}, true},
		{"fail_type_mismatch", map[string]any{"v": 42}, map[string]any{"v": "eq:hello"}, true},
		{"fail_case_sensitive", map[string]any{"v": "Hello"}, map[string]any{"v": "eq:hello"}, true},
		{"fail_no_params", map[string]any{"v": "hello"}, map[string]any{"v": "eq"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestNe() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_different_string", map[string]any{"v": "hello"}, map[string]any{"v": "ne:world"}, false},
		{"pass_different_int", map[string]any{"v": 42}, map[string]any{"v": "ne:43"}, false},
		{"pass_case_sensitive", map[string]any{"v": "Hello"}, map[string]any{"v": "ne:hello"}, false},
		{"pass_type_mismatch", map[string]any{"v": 42}, map[string]any{"v": "ne:hello"}, false},
		{"fail_same_string", map[string]any{"v": "hello"}, map[string]any{"v": "ne:hello"}, true},
		{"fail_same_int", map[string]any{"v": 42}, map[string]any{"v": "ne:42"}, true},
		{"fail_same_zero", map[string]any{"v": 0}, map[string]any{"v": "ne:0"}, true},
		{"fail_same_bool", map[string]any{"v": true}, map[string]any{"v": "ne:true"}, true},
		{"fail_no_params", map[string]any{"v": "hello"}, map[string]any{"v": "ne"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestIn() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_in_list", map[string]any{"x": "a"}, map[string]any{"x": "in:a,b,c"}, false},
		{"pass_last_in_list", map[string]any{"x": "c"}, map[string]any{"x": "in:a,b,c"}, false},
		{"pass_int_as_string", map[string]any{"x": 1}, map[string]any{"x": "in:1,2,3"}, false},
		{"pass_single_value", map[string]any{"x": "yes"}, map[string]any{"x": "in:yes"}, false},
		{"pass_empty_skipped", map[string]any{"x": ""}, map[string]any{"x": "in:a,b,c"}, false},
		{"pass_numeric_string_in_list", map[string]any{"status": "1"}, map[string]any{"status": "in:1,2,3"}, false},
		{"pass_integer_in_string_list", map[string]any{"status": 1}, map[string]any{"status": "in:1,2,3"}, false},
		{"pass_single_only", map[string]any{"x": "only"}, map[string]any{"x": "in:only"}, false},
		{"fail_not_in_list", map[string]any{"x": "d"}, map[string]any{"x": "in:a,b,c"}, true},
		{"fail_case_sensitive", map[string]any{"x": "A"}, map[string]any{"x": "in:a,b,c"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestNotIn() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_not_in_list", map[string]any{"x": "d"}, map[string]any{"x": "not_in:a,b,c"}, false},
		{"pass_case_different", map[string]any{"x": "A"}, map[string]any{"x": "not_in:a,b,c"}, false},
		{"pass_empty", map[string]any{"x": ""}, map[string]any{"x": "not_in:a,b,c"}, false},
		{"fail_in_list", map[string]any{"x": "a"}, map[string]any{"x": "not_in:a,b,c"}, true},
		{"fail_last_in_list", map[string]any{"x": "c"}, map[string]any{"x": "not_in:a,b,c"}, true},
		{"pass_pending_not_banned", map[string]any{"status": "pending"}, map[string]any{"status": "not_in:banned,deleted"}, false},
		{"fail_integer_match", map[string]any{"id": 0}, map[string]any{"id": "not_in:0,999"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestInArray() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_in_array", map[string]any{"x": "a", "arr": []any{"a", "b", "c"}}, map[string]any{"x": "in_array:arr"}, false},
		{"fail_not_in_array", map[string]any{"x": "d", "arr": []any{"a", "b", "c"}}, map[string]any{"x": "in_array:arr"}, true},
		{"fail_array_missing", map[string]any{"x": "a"}, map[string]any{"x": "in_array:arr"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestInArrayKeys() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_has_key", map[string]any{"x": map[string]any{"a": 1, "b": 2}}, map[string]any{"x": "in_array_keys:a,c"}, false},
		{"fail_no_key", map[string]any{"x": map[string]any{"d": 1}}, map[string]any{"x": "in_array_keys:a,b"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 10. Date Rules =====

func (s *RulesTestSuite) TestDate() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_date_only", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "date"}, false},
		{"pass_datetime_space", map[string]any{"x": "2024-01-15 10:30:00"}, map[string]any{"x": "date"}, false},
		{"pass_rfc3339", map[string]any{"x": "2024-01-15T10:30:00Z"}, map[string]any{"x": "date"}, false},
		{"pass_rfc3339_offset", map[string]any{"x": "2024-01-15T10:30:00+08:00"}, map[string]any{"x": "date"}, false},
		{"pass_datetime_t", map[string]any{"x": "2024-01-15T10:30:00"}, map[string]any{"x": "date"}, false},
		{"pass_empty_skipped", map[string]any{"x": ""}, map[string]any{"x": "date"}, false},
		{"fail_invalid", map[string]any{"x": "not-a-date"}, map[string]any{"x": "date"}, true},
		{"fail_numbers_only", map[string]any{"x": "20240115"}, map[string]any{"x": "date"}, true},
		{"fail_partial_date", map[string]any{"x": "2024-01"}, map[string]any{"x": "date"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDateFormat() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_date_format", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "date_format:2006-01-02"}, false},
		{"pass_datetime_format", map[string]any{"x": "2024-01-15 10:30:00"}, map[string]any{"x": "date_format:2006-01-02 15:04:05"}, false},
		{"pass_custom_format", map[string]any{"x": "15/01/2024"}, map[string]any{"x": "date_format:02/01/2006"}, false},
		{"pass_time_only", map[string]any{"x": "10:30:00"}, map[string]any{"x": "date_format:15:04:05"}, false},
		{"fail_wrong_format", map[string]any{"x": "15/01/2024"}, map[string]any{"x": "date_format:2006-01-02"}, true},
		{"fail_invalid_date", map[string]any{"x": "not-a-date"}, map[string]any{"x": "date_format:2006-01-02"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDateEquals() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_equal", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "date_equals:2024-01-15"}, false},
		{"pass_equal_datetime", map[string]any{"x": "2024-01-15T10:30:00Z"}, map[string]any{"x": "date_equals:2024-01-15T10:30:00Z"}, false},
		{"fail_not_equal", map[string]any{"x": "2024-01-16"}, map[string]any{"x": "date_equals:2024-01-15"}, true},
		{"fail_invalid_date", map[string]any{"x": "not-a-date"}, map[string]any{"x": "date_equals:2024-01-15"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestBefore() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_before", map[string]any{"x": "2024-01-14"}, map[string]any{"x": "before:2024-01-15"}, false},
		{"fail_after", map[string]any{"x": "2024-01-16"}, map[string]any{"x": "before:2024-01-15"}, true},
		{"fail_equal", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "before:2024-01-15"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestBeforeOrEqual() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_before", map[string]any{"x": "2024-01-14"}, map[string]any{"x": "before_or_equal:2024-01-15"}, false},
		{"pass_equal", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "before_or_equal:2024-01-15"}, false},
		{"fail_after", map[string]any{"x": "2024-01-16"}, map[string]any{"x": "before_or_equal:2024-01-15"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAfter() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_after", map[string]any{"x": "2024-01-16"}, map[string]any{"x": "after:2024-01-15"}, false},
		{"fail_before", map[string]any{"x": "2024-01-14"}, map[string]any{"x": "after:2024-01-15"}, true},
		{"fail_equal", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "after:2024-01-15"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestAfterOrEqual() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_after", map[string]any{"x": "2024-01-16"}, map[string]any{"x": "after_or_equal:2024-01-15"}, false},
		{"pass_equal", map[string]any{"x": "2024-01-15"}, map[string]any{"x": "after_or_equal:2024-01-15"}, false},
		{"fail_before", map[string]any{"x": "2024-01-14"}, map[string]any{"x": "after_or_equal:2024-01-15"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestTimezone() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_utc", map[string]any{"x": "UTC"}, map[string]any{"x": "timezone"}, false},
		{"pass_named", map[string]any{"x": "America/New_York"}, map[string]any{"x": "timezone"}, false},
		{"pass_asia", map[string]any{"x": "Asia/Shanghai"}, map[string]any{"x": "timezone"}, false},
		{"pass_europe", map[string]any{"x": "Europe/London"}, map[string]any{"x": "timezone"}, false},
		{"pass_local", map[string]any{"x": "Local"}, map[string]any{"x": "timezone"}, false},
		{"pass_empty_skipped", map[string]any{"x": ""}, map[string]any{"x": "timezone"}, false},
		{"pass_abbreviation_est", map[string]any{"x": "EST"}, map[string]any{"x": "timezone"}, false},
		{"fail_invalid", map[string]any{"x": "Not/A/Timezone"}, map[string]any{"x": "timezone"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestDateWithFieldReference() {
	s.Run("before_field_reference", func() {
		v := s.makeValidator(
			map[string]any{"start": "2024-01-10", "end": "2024-01-15"},
			map[string]any{"start": "before:end"},
		)
		s.False(v.Fails())
	})

	s.Run("after_field_reference", func() {
		v := s.makeValidator(
			map[string]any{"start": "2024-01-10", "end": "2024-01-15"},
			map[string]any{"end": "after:start"},
		)
		s.False(v.Fails())
	})
}

// ===== 11. Exclude Rules =====

func (s *RulesTestSuite) TestExclude() {
	s.Run("excluded_from_validated", func() {
		v := s.makeValidator(
			map[string]any{"name": "go", "secret": "hidden"},
			map[string]any{"name": "required", "secret": "exclude"},
		)
		s.False(v.Fails())
		_, exists := v.Validated()["secret"]
		s.False(exists)
		s.Equal("go", v.Validated()["name"])
	})
}

func (s *RulesTestSuite) TestExcludeIf() {
	s.Run("excluded_when_condition_met", func() {
		v := s.makeValidator(
			map[string]any{"type": "free", "cc": "1234"},
			map[string]any{"type": "required", "cc": "exclude_if:type,free"},
		)
		s.False(v.Fails())
		_, exists := v.Validated()["cc"]
		s.False(exists)
	})

	s.Run("not_excluded_when_condition_not_met", func() {
		v := s.makeValidator(
			map[string]any{"type": "paid", "cc": "1234"},
			map[string]any{"type": "required", "cc": "exclude_if:type,free"},
		)
		s.False(v.Fails())
		s.Equal("1234", v.Validated()["cc"])
	})
}

func (s *RulesTestSuite) TestExcludeUnless() {
	s.Run("excluded_when_condition_not_met", func() {
		v := s.makeValidator(
			map[string]any{"role": "user", "admin_note": "note"},
			map[string]any{"role": "required", "admin_note": "exclude_unless:role,admin"},
		)
		s.False(v.Fails())
		_, exists := v.Validated()["admin_note"]
		s.False(exists)
	})

	s.Run("not_excluded_when_condition_met", func() {
		v := s.makeValidator(
			map[string]any{"role": "admin", "admin_note": "note"},
			map[string]any{"role": "required", "admin_note": "exclude_unless:role,admin"},
		)
		s.False(v.Fails())
		s.Equal("note", v.Validated()["admin_note"])
	})
}

func (s *RulesTestSuite) TestExcludeWith() {
	s.Run("excluded_when_other_present", func() {
		v := s.makeValidator(
			map[string]any{"email": "a@b.com", "phone": "123"},
			map[string]any{"email": "required", "phone": "exclude_with:email"},
		)
		s.False(v.Fails())
		_, exists := v.Validated()["phone"]
		s.False(exists)
	})
}

func (s *RulesTestSuite) TestExcludeWithout() {
	s.Run("excluded_when_other_absent", func() {
		v := s.makeValidator(
			map[string]any{"phone": "123"},
			map[string]any{"phone": "required", "note": "exclude_without:email"},
		)
		s.False(v.Fails())
	})
}

// ===== 12. File Rules =====

func (s *RulesTestSuite) TestFile() {
	fh := &multipart.FileHeader{Filename: "test.txt", Size: 100}
	fh2 := &multipart.FileHeader{Filename: "test2.txt", Size: 200}
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_file", map[string]any{"x": fh}, map[string]any{"x": "file"}, false},
		{"pass_multiple_files", map[string]any{"x": []*multipart.FileHeader{fh, fh2}}, map[string]any{"x": "file"}, false},
		{"fail_string", map[string]any{"x": "not-a-file"}, map[string]any{"x": "file"}, true},
		{"fail_empty_slice", map[string]any{"x": []*multipart.FileHeader{}}, map[string]any{"x": "file"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestImage() {
	// JPEG: minimal valid JFIF
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
	// PNG: minimal valid PNG header
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	// PDF: minimal PDF header
	pdfData := []byte{0x25, 0x50, 0x44, 0x46, 0x2D}

	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_jpeg", map[string]any{"x": makeFileHeader(s.T(), "img.jpg", jpegData)}, map[string]any{"x": "image"}, false},
		{"pass_png", map[string]any{"x": makeFileHeader(s.T(), "img.png", pngData)}, map[string]any{"x": "image"}, false},
		{"pass_multiple_images", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.jpg", jpegData),
			makeFileHeader(s.T(), "b.png", pngData),
		}}, map[string]any{"x": "image"}, false},
		{"fail_pdf", map[string]any{"x": makeFileHeader(s.T(), "doc.pdf", pdfData)}, map[string]any{"x": "image"}, true},
		{"fail_multiple_one_not_image", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.jpg", jpegData),
			makeFileHeader(s.T(), "doc.pdf", pdfData),
		}}, map[string]any{"x": "image"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMimes() {
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	gifData := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}

	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_png", map[string]any{"x": makeFileHeader(s.T(), "photo.png", pngData)}, map[string]any{"x": "mimes:png,jpg"}, false},
		{"pass_multiple_pngs", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.png", pngData),
			makeFileHeader(s.T(), "b.png", pngData),
		}}, map[string]any{"x": "mimes:png"}, false},
		{"fail_gif", map[string]any{"x": makeFileHeader(s.T(), "photo.gif", gifData)}, map[string]any{"x": "mimes:jpg,png"}, true},
		{"fail_multiple_one_mismatch", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.png", pngData),
			makeFileHeader(s.T(), "b.jpg", jpegData),
		}}, map[string]any{"x": "mimes:png"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMimetypes() {
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	pdfData := []byte{0x25, 0x50, 0x44, 0x46, 0x2D}
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}

	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_match", map[string]any{"x": makeFileHeader(s.T(), "img.png", pngData)}, map[string]any{"x": "mimetypes:image/png,image/jpeg"}, false},
		{"pass_multiple_match", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.png", pngData),
			makeFileHeader(s.T(), "b.jpg", jpegData),
		}}, map[string]any{"x": "mimetypes:image/png,image/jpeg"}, false},
		{"fail_no_match", map[string]any{"x": makeFileHeader(s.T(), "doc.pdf", pdfData)}, map[string]any{"x": "mimetypes:image/png,image/jpeg"}, true},
		{"fail_multiple_one_mismatch", map[string]any{"x": []*multipart.FileHeader{
			makeFileHeader(s.T(), "a.png", pngData),
			makeFileHeader(s.T(), "doc.pdf", pdfData),
		}}, map[string]any{"x": "mimetypes:image/png,image/jpeg"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestExtensions() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_match", map[string]any{"x": &multipart.FileHeader{Filename: "doc.pdf"}}, map[string]any{"x": "extensions:pdf,docx"}, false},
		{"pass_multiple_match", map[string]any{"x": []*multipart.FileHeader{
			{Filename: "a.pdf"},
			{Filename: "b.docx"},
		}}, map[string]any{"x": "extensions:pdf,docx"}, false},
		{"fail_no_match", map[string]any{"x": &multipart.FileHeader{Filename: "doc.txt"}}, map[string]any{"x": "extensions:pdf,docx"}, true},
		{"fail_multiple_one_mismatch", map[string]any{"x": []*multipart.FileHeader{
			{Filename: "a.pdf"},
			{Filename: "b.txt"},
		}}, map[string]any{"x": "extensions:pdf,docx"}, true},
		{"fail_empty_slice", map[string]any{"x": []*multipart.FileHeader{}}, map[string]any{"x": "extensions:pdf"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 13. Control Rules =====

func (s *RulesTestSuite) TestBail() {
	s.Run("stops_on_first_error", func() {
		v := s.makeValidator(
			map[string]any{"email": ""},
			map[string]any{"email": "bail|required|email"},
		)
		s.True(v.Fails())
		errs := v.Errors().Get("email")
		s.Len(errs, 1)
	})

	s.Run("reports_all_errors_without_bail", func() {
		v := s.makeValidator(
			map[string]any{"x": ""},
			map[string]any{"x": "required|email"},
		)
		s.True(v.Fails())
		// Without bail, required is an implicit rule failure that triggers
		errs := v.Errors().Get("x")
		s.GreaterOrEqual(len(errs), 1)
	})

	s.Run("passes_when_valid_with_bail", func() {
		v := s.makeValidator(
			map[string]any{"email": "user@example.com"},
			map[string]any{"email": "bail|required|email"},
		)
		s.False(v.Fails())
	})
}

func (s *RulesTestSuite) TestNullable() {
	s.Run("allows_nil", func() {
		v := s.makeValidator(
			map[string]any{"name": "go"},
			map[string]any{"name": "required", "email": "nullable|email"},
		)
		s.False(v.Fails())
	})

	s.Run("validates_when_present", func() {
		v := s.makeValidator(
			map[string]any{"email": "bad"},
			map[string]any{"email": "nullable|email"},
		)
		s.True(v.Fails())
	})

	s.Run("allows_nil_value", func() {
		v := s.makeValidator(
			map[string]any{"name": nil},
			map[string]any{"name": "nullable|string"},
		)
		s.False(v.Fails())
	})

	s.Run("passes_with_valid_value", func() {
		v := s.makeValidator(
			map[string]any{"name": "hello"},
			map[string]any{"name": "nullable|string|min:3"},
		)
		s.False(v.Fails())
	})
}

func (s *RulesTestSuite) TestSometimes() {
	s.Run("skips_when_absent", func() {
		v := s.makeValidator(
			map[string]any{"name": "go"},
			map[string]any{"name": "required", "email": "sometimes|required|email"},
		)
		s.False(v.Fails())
	})

	s.Run("validates_when_present", func() {
		v := s.makeValidator(
			map[string]any{"name": "go", "email": "bad"},
			map[string]any{"name": "required", "email": "sometimes|required|email"},
		)
		s.True(v.Fails())
	})

	s.Run("passes_when_present_and_valid", func() {
		v := s.makeValidator(
			map[string]any{"name": "go", "email": "user@example.com"},
			map[string]any{"name": "required", "email": "sometimes|required|email"},
		)
		s.False(v.Fails())
	})
}

// ===== 14. Other Rules =====

func (s *RulesTestSuite) TestRequiredArrayKeys() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_all_keys", map[string]any{"x": map[string]any{"a": 1, "b": 2}}, map[string]any{"x": "required_array_keys:a,b"}, false},
		{"fail_missing_key", map[string]any{"x": map[string]any{"a": 1}}, map[string]any{"x": "required_array_keys:a,b"}, true},
		{"fail_not_map", map[string]any{"x": "hello"}, map[string]any{"x": "required_array_keys:a"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestEncoding() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_utf8", map[string]any{"x": "hello"}, map[string]any{"x": "encoding:utf-8"}, false},
		{"pass_ascii_encoding", map[string]any{"x": "hello"}, map[string]any{"x": "encoding:ascii"}, false},
		{"fail_non_ascii", map[string]any{"x": "héllo"}, map[string]any{"x": "encoding:ascii"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== 15. Engine Feature Tests =====

func (s *RulesTestSuite) TestWildcardExpansion() {
	s.Run("validates_all_items", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{
				map[string]any{"name": "a"},
				map[string]any{"name": ""},
			}},
			map[string]any{"items.*.name": "required"},
		)
		s.True(v.Fails())
	})

	s.Run("all_valid", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{
				map[string]any{"name": "a"},
				map[string]any{"name": "b"},
			}},
			map[string]any{"items.*.name": "required|string"},
		)
		s.False(v.Fails())
	})
}

func (s *RulesTestSuite) TestNestedDotNotation() {
	s.Run("nested_access", func() {
		v := s.makeValidator(
			map[string]any{"user": map[string]any{"profile": map[string]any{"name": "Go"}}},
			map[string]any{"user.profile.name": "required|string|min:2"},
		)
		s.False(v.Fails())
	})

	s.Run("nested_missing", func() {
		v := s.makeValidator(
			map[string]any{"user": map[string]any{"profile": map[string]any{}}},
			map[string]any{"user.profile.name": "required"},
		)
		s.True(v.Fails())
	})
}

func (s *RulesTestSuite) TestCustomMessages() {
	s.Run("field_rule_message", func() {
		v := s.makeValidator(
			map[string]any{"name": ""},
			map[string]any{"name": "required"},
			Messages(map[string]string{"name.required": "Name is mandatory."}),
		)
		s.True(v.Fails())
		s.Equal("Name is mandatory.", v.Errors().One("name"))
	})
}

func (s *RulesTestSuite) TestCustomAttributes() {
	s.Run("attribute_replacement", func() {
		v := s.makeValidator(
			map[string]any{"email_address": ""},
			map[string]any{"email_address": "required"},
			Attributes(map[string]string{"email_address": "Email"}),
		)
		s.True(v.Fails())
		s.Equal("The Email field is required.", v.Errors().One("email_address"))
	})
}

func (s *RulesTestSuite) TestValidatedOnlyReturnsRuledFields() {
	s.Run("excludes_extra_fields", func() {
		v := s.makeValidator(
			map[string]any{"name": "go", "email": "a@b.com", "extra": "x"},
			map[string]any{"name": "required", "email": "required|email"},
		)
		s.False(v.Fails())
		validated := v.Validated()
		s.Equal("go", validated["name"])
		s.Equal("a@b.com", validated["email"])
		_, exists := validated["extra"]
		s.False(exists)
	})
}

func (s *RulesTestSuite) TestMultiRuleCombination() {
	s.Run("all_pass", func() {
		v := s.makeValidator(
			map[string]any{"name": "goravel"},
			map[string]any{"name": "required|string|min:3|max:50"},
		)
		s.False(v.Fails())
	})

	s.Run("min_fails", func() {
		v := s.makeValidator(
			map[string]any{"name": "go"},
			map[string]any{"name": "required|string|min:3|max:50"},
		)
		s.True(v.Fails())
		errs := v.Errors().Get("name")
		s.Contains(errs, "min")
	})

	s.Run("required_string_email_min", func() {
		v := s.makeValidator(
			map[string]any{"email": "a@b.com"},
			map[string]any{"email": "required|string|email|min:5"},
		)
		s.False(v.Fails())
	})

	s.Run("required_string_email_min_fail_too_short", func() {
		v := s.makeValidator(
			map[string]any{"email": "a@b"},
			map[string]any{"email": "required|string|email|min:5"},
		)
		s.True(v.Fails())
	})

	s.Run("required_integer_between", func() {
		v := s.makeValidator(
			map[string]any{"age": 25},
			map[string]any{"age": "required|integer|between:18,100"},
		)
		s.False(v.Fails())
	})

	s.Run("required_integer_between_fail_too_young", func() {
		v := s.makeValidator(
			map[string]any{"age": 10},
			map[string]any{"age": "required|integer|between:18,100"},
		)
		s.True(v.Fails())
	})

	s.Run("nullable_string_email", func() {
		v := s.makeValidator(
			map[string]any{"email": nil},
			map[string]any{"email": "nullable|string|email"},
		)
		s.False(v.Fails())
	})

	s.Run("sometimes_required_string", func() {
		v := s.makeValidator(
			map[string]any{},
			map[string]any{"nickname": "sometimes|required|string"},
		)
		s.False(v.Fails())
	})

	s.Run("sometimes_required_string_present_empty", func() {
		v := s.makeValidator(
			map[string]any{"nickname": ""},
			map[string]any{"nickname": "sometimes|required|string"},
		)
		s.True(v.Fails())
	})

	s.Run("array_min_max_combination", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{1, 2, 3}},
			map[string]any{"items": "required|array|min:1|max:5"},
		)
		s.False(v.Fails())
	})

	s.Run("string_alpha_dash_max", func() {
		v := s.makeValidator(
			map[string]any{"slug": "my-post_123"},
			map[string]any{"slug": "required|string|alpha_dash|max:50"},
		)
		s.False(v.Fails())
	})

	s.Run("numeric_multiple_of_between", func() {
		v := s.makeValidator(
			map[string]any{"qty": 15},
			map[string]any{"qty": "required|numeric|multiple_of:5|between:5,50"},
		)
		s.False(v.Fails())
	})

	s.Run("multiple_fields_all_pass", func() {
		v := s.makeValidator(
			map[string]any{
				"name":  "goravel",
				"email": "go@example.com",
				"age":   25,
				"role":  "admin",
			},
			map[string]any{
				"name":  "required|string|min:3|max:50",
				"email": "required|email",
				"age":   "required|integer|between:18,100",
				"role":  "required|in:admin,user,guest",
			},
		)
		s.False(v.Fails())
	})

	s.Run("multiple_fields_some_fail", func() {
		v := s.makeValidator(
			map[string]any{
				"name":  "go",
				"email": "bad",
				"age":   200,
				"role":  "superadmin",
			},
			map[string]any{
				"name":  "required|string|min:3|max:50",
				"email": "required|email",
				"age":   "required|integer|between:18,100",
				"role":  "required|in:admin,user,guest",
			},
		)
		s.True(v.Fails())
		errs := v.Errors().All()
		s.Contains(errs, "name")
		s.Contains(errs, "email")
		s.Contains(errs, "age")
		s.Contains(errs, "role")
	})
}

func (s *RulesTestSuite) TestDistinct() {
	s.Run("pass_distinct_values", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{"a", "b", "c"}},
			map[string]any{"items.*": "distinct"},
		)
		s.False(v.Fails())
	})

	s.Run("fail_duplicate_values", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{"a", "b", "a"}},
			map[string]any{"items.*": "distinct"},
		)
		s.True(v.Fails())
	})
}

// ===== Error Message Tests =====

func (s *RulesTestSuite) TestErrorMessages() {
	s.Run("required_message", func() {
		v := s.makeValidator(
			map[string]any{"name": ""},
			map[string]any{"name": "required"},
		)
		s.True(v.Fails())
		s.Equal("The name field is required.", v.Errors().One("name"))
	})

	s.Run("min_string_message", func() {
		v := s.makeValidator(
			map[string]any{"name": "ab"},
			map[string]any{"name": "string|min:3"},
		)
		s.True(v.Fails())
		s.Equal("The name field must be at least 3 characters.", v.Errors().One("name"))
	})

	s.Run("max_numeric_message", func() {
		v := s.makeValidator(
			map[string]any{"age": 200},
			map[string]any{"age": "numeric|max:150"},
		)
		s.True(v.Fails())
		s.Equal("The age field must not be greater than 150.", v.Errors().One("age"))
	})

	s.Run("between_string_message", func() {
		v := s.makeValidator(
			map[string]any{"x": "ab"},
			map[string]any{"x": "string|between:3,5"},
		)
		s.True(v.Fails())
		s.Equal("The x field must be between 3 and 5 characters.", v.Errors().One("x"))
	})

	s.Run("in_message", func() {
		v := s.makeValidator(
			map[string]any{"status": "bad"},
			map[string]any{"status": "in:active,inactive"},
		)
		s.True(v.Fails())
		s.Equal("The selected status is invalid.", v.Errors().One("status"))
	})

	s.Run("email_message", func() {
		v := s.makeValidator(
			map[string]any{"email": "bad"},
			map[string]any{"email": "email"},
		)
		s.True(v.Fails())
		s.Equal("The email field must be a valid email address.", v.Errors().One("email"))
	})

	s.Run("confirmed_message", func() {
		v := s.makeValidator(
			map[string]any{"pw": "a", "pw_confirmation": "b"},
			map[string]any{"pw": "confirmed"},
		)
		s.True(v.Fails())
		s.Equal("The pw field confirmation does not match.", v.Errors().One("pw"))
	})

	s.Run("same_message", func() {
		v := s.makeValidator(
			map[string]any{"a": "x", "b": "y"},
			map[string]any{"a": "same:b"},
		)
		s.True(v.Fails())
		s.Equal("The a field must match b.", v.Errors().One("a"))
	})

	s.Run("size_array_message", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{1, 2}},
			map[string]any{"items": "array|size:3"},
		)
		s.True(v.Fails())
		s.Equal("The items field must contain 3 items.", v.Errors().One("items"))
	})

	s.Run("uuid_message", func() {
		v := s.makeValidator(
			map[string]any{"id": "bad"},
			map[string]any{"id": "uuid"},
		)
		s.True(v.Fails())
		s.Equal("The id field must be a valid UUID.", v.Errors().One("id"))
	})

	s.Run("underscore_to_space_in_attribute", func() {
		v := s.makeValidator(
			map[string]any{"first_name": ""},
			map[string]any{"first_name": "required"},
		)
		s.True(v.Fails())
		s.Equal("The first name field is required.", v.Errors().One("first_name"))
	})

	s.Run("custom_message_for_specific_field_rule", func() {
		v := s.makeValidator(
			map[string]any{"email": "bad"},
			map[string]any{"email": "email"},
			Messages(map[string]string{"email.email": "Please enter a valid email"}),
		)
		s.True(v.Fails())
		s.Equal("Please enter a valid email", v.Errors().One("email"))
	})

	s.Run("custom_message_for_rule_only", func() {
		v := s.makeValidator(
			map[string]any{"email": "bad"},
			map[string]any{"email": "email"},
			Messages(map[string]string{"email": "Invalid email format"}),
		)
		s.True(v.Fails())
		s.Equal("Invalid email format", v.Errors().One("email"))
	})

	s.Run("custom_attribute_name", func() {
		v := s.makeValidator(
			map[string]any{"email_addr": ""},
			map[string]any{"email_addr": "required"},
			Attributes(map[string]string{"email_addr": "email address"}),
		)
		s.True(v.Fails())
		s.Equal("The email address field is required.", v.Errors().One("email_addr"))
	})

	s.Run("custom_message_with_attribute_placeholder", func() {
		v := s.makeValidator(
			map[string]any{"user_name": ""},
			map[string]any{"user_name": "required"},
			Messages(map[string]string{"user_name.required": ":attribute is mandatory"}),
			Attributes(map[string]string{"user_name": "Username"}),
		)
		s.True(v.Fails())
		s.Equal("Username is mandatory", v.Errors().One("user_name"))
	})

	s.Run("min_numeric_message", func() {
		v := s.makeValidator(
			map[string]any{"price": 5},
			map[string]any{"price": "numeric|min:10"},
		)
		s.True(v.Fails())
		s.Equal("The price field must be at least 10.", v.Errors().One("price"))
	})

	s.Run("max_array_message", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{1, 2, 3, 4, 5}},
			map[string]any{"items": "array|max:3"},
		)
		s.True(v.Fails())
		s.Equal("The items field must not have more than 3 items.", v.Errors().One("items"))
	})

	s.Run("between_numeric_message", func() {
		v := s.makeValidator(
			map[string]any{"score": 100},
			map[string]any{"score": "numeric|between:1,10"},
		)
		s.True(v.Fails())
		s.Equal("The score field must be between 1 and 10.", v.Errors().One("score"))
	})

	s.Run("digits_message", func() {
		v := s.makeValidator(
			map[string]any{"pin": "12"},
			map[string]any{"pin": "digits:4"},
		)
		s.True(v.Fails())
		s.Equal("The pin field must be 4 digits.", v.Errors().One("pin"))
	})

	s.Run("required_if_message", func() {
		v := s.makeValidator(
			map[string]any{"type": "admin"},
			map[string]any{"role": "required_if:type,admin"},
		)
		s.True(v.Fails())
		s.Equal("The role field is required when type is admin.", v.Errors().One("role"))
	})

	s.Run("prohibited_if_message", func() {
		v := s.makeValidator(
			map[string]any{"role": "guest", "token": "abc"},
			map[string]any{"token": "prohibited_if:role,guest"},
		)
		s.True(v.Fails())
		s.Equal("The token field is prohibited when role is guest.", v.Errors().One("token"))
	})

	s.Run("starts_with_message", func() {
		v := s.makeValidator(
			map[string]any{"code": "xyz_bad"},
			map[string]any{"code": "starts_with:abc_,def_"},
		)
		s.True(v.Fails())
		s.Contains(v.Errors().One("code"), "must start with one of the following")
	})

	s.Run("date_before_message", func() {
		v := s.makeValidator(
			map[string]any{"d": "2030-01-01"},
			map[string]any{"d": "before:2025-01-01"},
		)
		s.True(v.Fails())
		s.Equal("The d field must be a date before 2025-01-01.", v.Errors().One("d"))
	})

	s.Run("min_array_message", func() {
		v := s.makeValidator(
			map[string]any{"items": []any{1}},
			map[string]any{"items": "array|min:3"},
		)
		s.True(v.Fails())
		s.Equal("The items field must have at least 3 items.", v.Errors().One("items"))
	})

	s.Run("size_numeric_message", func() {
		v := s.makeValidator(
			map[string]any{"qty": 5},
			map[string]any{"qty": "numeric|size:10"},
		)
		s.True(v.Fails())
		s.Equal("The qty field must be 10.", v.Errors().One("qty"))
	})

	s.Run("multiple_of_message", func() {
		v := s.makeValidator(
			map[string]any{"n": 7},
			map[string]any{"n": "multiple_of:3"},
		)
		s.True(v.Fails())
		s.Equal("The n field must be a multiple of 3.", v.Errors().One("n"))
	})
}

// ===== Rule Alias Tests =====

func (s *RulesTestSuite) TestIntAlias() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_int_alias", map[string]any{"n": 42}, map[string]any{"n": "int"}, false},
		{"pass_string_int_alias", map[string]any{"n": "42"}, map[string]any{"n": "int"}, false},
		{"fail_float_int_alias", map[string]any{"n": 3.14}, map[string]any{"n": "int"}, true},
		{"fail_string_int_alias", map[string]any{"n": "abc"}, map[string]any{"n": "int"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestBoolAlias() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_bool_alias_true", map[string]any{"ok": true}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_false", map[string]any{"ok": false}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_1", map[string]any{"ok": 1}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_0", map[string]any{"ok": 0}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_true", map[string]any{"ok": "true"}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_false", map[string]any{"ok": "false"}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_on", map[string]any{"ok": "on"}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_off", map[string]any{"ok": "off"}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_yes", map[string]any{"ok": "yes"}, map[string]any{"ok": "bool"}, false},
		{"pass_bool_alias_string_no", map[string]any{"ok": "no"}, map[string]any{"ok": "bool"}, false},
		{"fail_bool_alias_string", map[string]any{"ok": "abc"}, map[string]any{"ok": "bool"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestSliceAlias() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_slice_alias", map[string]any{"items": []any{1, 2}}, map[string]any{"items": "slice"}, false},
		{"fail_slice_alias_map", map[string]any{"items": map[string]any{"a": 1}}, map[string]any{"items": "slice"}, true},
		{"fail_slice_alias_string", map[string]any{"items": "abc"}, map[string]any{"items": "slice"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

func (s *RulesTestSuite) TestMacAlias() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_mac_alias", map[string]any{"mac": "00:1A:2B:3C:4D:5E"}, map[string]any{"mac": "mac"}, false},
		{"fail_mac_alias", map[string]any{"mac": "invalid"}, map[string]any{"mac": "mac"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== Active URL Tests =====

func (s *RulesTestSuite) TestActiveUrl() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		{"pass_goravel", map[string]any{"u": "https://goravel.dev"}, map[string]any{"u": "active_url"}, false},
		{"fail_non_string", map[string]any{"u": 123}, map[string]any{"u": "active_url"}, true},
		{"fail_no_host", map[string]any{"u": "not-a-url"}, map[string]any{"u": "active_url"}, true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}

// ===== Array Syntax ([]string) Tests =====

func (s *RulesTestSuite) TestArraySyntaxBasic() {
	s.Run("string_and_array_syntax_equivalent", func() {
		data := map[string]any{"name": "goravel", "age": 18}
		v1 := s.makeValidator(data, map[string]any{
			"name": "required|string|min:3",
			"age":  "required|integer",
		})
		v2 := s.makeValidator(data, map[string]any{
			"name": []string{"required", "string", "min:3"},
			"age":  []string{"required", "integer"},
		})
		s.Equal(v1.Fails(), v2.Fails())
	})

	s.Run("array_syntax_regex_with_pipe", func() {
		v := s.makeValidator(
			map[string]any{"code": "foo"},
			map[string]any{"code": []string{"required", "regex:^(foo|bar)$"}},
		)
		s.False(v.Fails())
	})

	s.Run("array_syntax_regex_with_pipe_fail", func() {
		v := s.makeValidator(
			map[string]any{"code": "baz"},
			map[string]any{"code": []string{"required", "regex:^(foo|bar)$"}},
		)
		s.True(v.Fails())
	})

	s.Run("array_syntax_regex_followed_by_more_rules", func() {
		v := s.makeValidator(
			map[string]any{"code": "foo"},
			map[string]any{"code": []string{"required", "regex:^(foo|bar)$", "string", "min:2"}},
		)
		s.False(v.Fails())
	})

	s.Run("array_syntax_not_regex_with_pipe", func() {
		v := s.makeValidator(
			map[string]any{"code": "baz"},
			map[string]any{"code": []string{"required", "not_regex:^(foo|bar)$"}},
		)
		s.False(v.Fails())
	})

	s.Run("array_syntax_mixed_string_and_array", func() {
		v := s.makeValidator(
			map[string]any{"name": "goravel", "code": "foo"},
			map[string]any{
				"name": "required|string",
				"code": []string{"required", "regex:^(foo|bar)$"},
			},
		)
		s.False(v.Fails())
	})

	s.Run("array_syntax_multiple_rules_fail", func() {
		v := s.makeValidator(
			map[string]any{"name": "ab"},
			map[string]any{"name": []string{"required", "string", "min:5", "max:10"}},
		)
		s.True(v.Fails())
	})

	s.Run("array_syntax_with_bail", func() {
		v := s.makeValidator(
			map[string]any{"email": ""},
			map[string]any{"email": []string{"bail", "required", "email"}},
		)
		s.True(v.Fails())
		// With bail, should only have the required error
		errs := v.Errors().Get("email")
		s.Equal(1, len(errs))
	})

	s.Run("array_syntax_empty_strings_ignored", func() {
		v := s.makeValidator(
			map[string]any{"x": "hello"},
			map[string]any{"x": []string{"", "required", "", "string"}},
		)
		s.False(v.Fails())
	})
}

// ===== Validated Data Tests =====

func (s *RulesTestSuite) TestValidatedDataExcludesUnruledFields() {
	v := s.makeValidator(
		map[string]any{"name": "go", "age": 25, "extra": "ignored"},
		map[string]any{"name": "required", "age": "required|integer"},
	)
	s.False(v.Fails())
	data := v.Validated()
	s.Equal(map[string]any{"name": "go", "age": 25}, data)
	_, hasExtra := data["extra"]
	s.False(hasExtra)
}

func (s *RulesTestSuite) TestValidatedDataWithExclusion() {
	s.Run("exclude_removes_field", func() {
		v := s.makeValidator(
			map[string]any{"name": "go", "secret": "hidden"},
			map[string]any{"name": "required", "secret": "exclude"},
		)
		s.False(v.Fails())
		data := v.Validated()
		_, hasSecret := data["secret"]
		s.False(hasSecret)
		s.Equal("go", data["name"])
	})

	s.Run("exclude_if_conditional", func() {
		v := s.makeValidator(
			map[string]any{"role": "admin", "token": "abc"},
			map[string]any{"role": "required", "token": "exclude_if:role,admin"},
		)
		s.False(v.Fails())
		data := v.Validated()
		_, hasToken := data["token"]
		s.False(hasToken)
	})

	s.Run("exclude_if_not_triggered", func() {
		v := s.makeValidator(
			map[string]any{"role": "user", "token": "abc"},
			map[string]any{"role": "required", "token": "exclude_if:role,admin"},
		)
		s.False(v.Fails())
		data := v.Validated()
		s.Equal("abc", data["token"])
	})
}

func (s *RulesTestSuite) TestValidatedDataNestedDot() {
	v := s.makeValidator(
		map[string]any{"user": map[string]any{"name": "go", "email": "a@b.com"}},
		map[string]any{"user.name": "required|string", "user.email": "required|email"},
	)
	s.False(v.Fails())
	data := v.Validated()
	user, ok := data["user"].(map[string]any)
	s.True(ok)
	s.Equal("go", user["name"])
	s.Equal("a@b.com", user["email"])
}

func (s *RulesTestSuite) TestValidatedDataWithWildcard() {
	s.Run("any slice", func() {
		v := s.makeValidator(
			map[string]any{"tags": []any{"go", "rust", "zig"}},
			map[string]any{"tags.*": "required|string"},
		)
		s.False(v.Fails())
		data := v.Validated()
		tags, ok := data["tags"].([]any)
		s.True(ok)
		s.Equal([]any{"go", "rust", "zig"}, tags)
	})

	s.Run("typed int slice", func() {
		v := s.makeValidator(
			map[string]any{"scores": []int{1, 2}},
			map[string]any{"scores.*": "required|integer"},
		)
		s.False(v.Fails())
		data := v.Validated()
		scores, ok := data["scores"].([]int)
		s.True(ok)
		s.Equal([]int{1, 2}, scores)
	})

	s.Run("nested wildcard", func() {
		v := s.makeValidator(
			map[string]any{
				"users": []any{
					map[string]any{"name": "alice", "email": "alice@example.com"},
					map[string]any{"name": "bob", "email": "bob@example.com"},
				},
			},
			map[string]any{"users.*.name": "required|string"},
		)
		s.False(v.Fails())
		data := v.Validated()
		users, ok := data["users"].([]any)
		s.True(ok)
		s.Equal(
			[]any{
				map[string]any{"name": "alice"},
				map[string]any{"name": "bob"},
			},
			users,
		)
	})
}

// ===== Error Bag Methods Tests =====

func (s *RulesTestSuite) TestErrorBagMethods() {
	s.Run("has_returns_false_when_no_errors", func() {
		v := s.makeValidator(
			map[string]any{"name": "go"},
			map[string]any{"name": "required"},
		)
		s.False(v.Fails())
		s.Nil(v.Errors())
	})

	s.Run("get_returns_all_errors_for_field", func() {
		v := s.makeValidator(
			map[string]any{"name": ""},
			map[string]any{"name": "required"},
		)
		s.True(v.Fails())
		errs := v.Errors().Get("name")
		s.NotNil(errs)
		s.Contains(errs, "required")
	})

	s.Run("all_returns_errors_for_all_fields", func() {
		v := s.makeValidator(
			map[string]any{},
			map[string]any{"a": "required", "b": "required"},
		)
		s.True(v.Fails())
		all := v.Errors().All()
		s.Contains(all, "a")
		s.Contains(all, "b")
	})

	s.Run("one_returns_first_error", func() {
		v := s.makeValidator(
			map[string]any{"name": ""},
			map[string]any{"name": "required"},
		)
		s.True(v.Fails())
		s.NotEmpty(v.Errors().One())
	})

	s.Run("one_with_field_returns_specific", func() {
		v := s.makeValidator(
			map[string]any{},
			map[string]any{"a": "required", "b": "required"},
		)
		s.True(v.Fails())
		s.NotEmpty(v.Errors().One("a"))
		s.NotEmpty(v.Errors().One("b"))
	})

	s.Run("has_returns_true_for_failed_field", func() {
		v := s.makeValidator(
			map[string]any{},
			map[string]any{"x": "required", "y": "required"},
		)
		s.True(v.Fails())
		s.True(v.Errors().Has("x"))
		s.True(v.Errors().Has("y"))
		s.False(v.Errors().Has("z"))
	})
}

func (s *RulesTestSuite) TestWildcardWithMultipleRules() {
	s.Run("wildcard_required_and_string", func() {
		v := s.makeValidator(
			map[string]any{"tags": []any{"go", "rust"}},
			map[string]any{"tags.*": "required|string"},
		)
		s.False(v.Fails())
	})

	s.Run("wildcard_fail_one_element", func() {
		v := s.makeValidator(
			map[string]any{"tags": []any{"go", ""}},
			map[string]any{"tags.*": "required"},
		)
		s.True(v.Fails())
	})

	s.Run("wildcard_with_min", func() {
		v := s.makeValidator(
			map[string]any{"names": []any{"ab", "cdef"}},
			map[string]any{"names.*": "string|min:3"},
		)
		s.True(v.Fails())
	})

	s.Run("wildcard_nested_objects", func() {
		v := s.makeValidator(
			map[string]any{
				"users": []any{
					map[string]any{"name": "alice"},
					map[string]any{"name": "bob"},
				},
			},
			map[string]any{"users.*.name": "required|string"},
		)
		s.False(v.Fails())
	})

	s.Run("wildcard_nested_fail", func() {
		v := s.makeValidator(
			map[string]any{
				"users": []any{
					map[string]any{"name": "alice"},
					map[string]any{"name": ""},
				},
			},
			map[string]any{"users.*.name": "required"},
		)
		s.True(v.Fails())
	})
}

func (s *RulesTestSuite) TestDeepNestedDotNotation() {
	s.Run("three_level_deep", func() {
		v := s.makeValidator(
			map[string]any{
				"config": map[string]any{
					"db": map[string]any{
						"host": "localhost",
					},
				},
			},
			map[string]any{"config.db.host": "required|string"},
		)
		s.False(v.Fails())
	})

	s.Run("three_level_deep_fail", func() {
		v := s.makeValidator(
			map[string]any{
				"config": map[string]any{
					"db": map[string]any{
						"host": "",
					},
				},
			},
			map[string]any{"config.db.host": "required"},
		)
		s.True(v.Fails())
	})

	s.Run("nested_with_multiple_fields", func() {
		v := s.makeValidator(
			map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			map[string]any{
				"server.host": "required|string",
				"server.port": "required|integer",
			},
		)
		s.False(v.Fails())
	})
}

// makeFileHeader creates a real multipart.FileHeader from file content so that
// detectMIME (which calls fh.Open()) works correctly in tests.
func makeFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	r := multipart.NewReader(&buf, w.Boundary())
	form, err := r.ReadForm(32 << 20)
	if err != nil {
		t.Fatal(err)
	}
	return form.File["file"][0]
}

// ---- Database Rules Tests ----

type DBRulesTestSuite struct {
	suite.Suite
	mockOrm   *mocksorm.Orm
	mockQuery *mocksorm.Query
}

func TestDBRulesTestSuite(t *testing.T) {
	suite.Run(t, new(DBRulesTestSuite))
}

func (s *DBRulesTestSuite) SetupTest() {
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockQuery = mocksorm.NewQuery(s.T())
	ormFacade = s.mockOrm
}

func (s *DBRulesTestSuite) TearDownTest() {
	ormFacade = nil
}

// --- exists rule tests ---

func (s *DBRulesTestSuite) TestRuleExists_SingleColumn_Found() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_SingleColumn_NotFound() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "notfound@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(false, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "notfound@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_DefaultColumn() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_MultipleColumns_OR() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("username", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "username"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_MultipleColumns_ThreeFields() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("username", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("phone", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(false, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "value",
		Parameters: []string{"users", "email", "username", "phone"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_ConnectionTable() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Connection("mysql").Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"mysql.users", "email"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_OrmNil() {
	ormFacade = nil

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_NoParameters() {
	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{},
	}
	s.False(ruleExists(ctx))
}

// --- unique rule tests ---

func (s *DBRulesTestSuite) TestRuleUnique_IsUnique() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_NotUnique() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "taken@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(1), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "taken@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_DefaultColumn() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithExcept() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "id", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithCustomIdColumnAndExcept() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("user_id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "user_id", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithMultipleExcepts() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"1", "2", "3"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "id", "1", "2", "3"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithDefaultIdColumn() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_ConnectionTable() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Connection("pgsql").Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"pgsql.users", "email"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_OrmNil() {
	ormFacade = nil

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_NoParameters() {
	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{},
	}
	s.False(ruleUnique(ctx))
}

// ===== Deprecated Rule Aliases =====

func (s *RulesTestSuite) TestDeprecatedAliases() {
	tests := []struct {
		name  string
		data  map[string]any
		rules map[string]any
		fails bool
	}{
		// len → size
		{"len_string_pass", map[string]any{"name": "hello"}, map[string]any{"name": "string|len:5"}, false},
		{"len_string_fail", map[string]any{"name": "hi"}, map[string]any{"name": "string|len:5"}, true},

		// min_len → min
		{"min_len_pass", map[string]any{"name": "hello"}, map[string]any{"name": "string|min_len:3"}, false},
		{"min_len_fail", map[string]any{"name": "hi"}, map[string]any{"name": "string|min_len:3"}, true},

		// max_len → max
		{"max_len_pass", map[string]any{"name": "hi"}, map[string]any{"name": "string|max_len:5"}, false},
		{"max_len_fail", map[string]any{"name": "hello world"}, map[string]any{"name": "string|max_len:5"}, true},

		// eq_field → same
		{"eq_field_pass", map[string]any{"a": "x", "b": "x"}, map[string]any{"a": "eq_field:b"}, false},
		{"eq_field_fail", map[string]any{"a": "x", "b": "y"}, map[string]any{"a": "eq_field:b"}, true},

		// ne_field → different
		{"ne_field_pass", map[string]any{"a": "x", "b": "y"}, map[string]any{"a": "ne_field:b"}, false},
		{"ne_field_fail", map[string]any{"a": "x", "b": "x"}, map[string]any{"a": "ne_field:b"}, true},

		// gt_field → gt
		{"gt_field_pass", map[string]any{"a": 10, "b": 5}, map[string]any{"a": "numeric|gt_field:b"}, false},
		{"gt_field_fail", map[string]any{"a": 3, "b": 5}, map[string]any{"a": "numeric|gt_field:b"}, true},

		// gte_field → gte
		{"gte_field_pass", map[string]any{"a": 5, "b": 5}, map[string]any{"a": "numeric|gte_field:b"}, false},
		{"gte_field_fail", map[string]any{"a": 3, "b": 5}, map[string]any{"a": "numeric|gte_field:b"}, true},

		// lt_field → lt
		{"lt_field_pass", map[string]any{"a": 3, "b": 5}, map[string]any{"a": "numeric|lt_field:b"}, false},
		{"lt_field_fail", map[string]any{"a": 10, "b": 5}, map[string]any{"a": "numeric|lt_field:b"}, true},

		// lte_field → lte
		{"lte_field_pass", map[string]any{"a": 5, "b": 5}, map[string]any{"a": "numeric|lte_field:b"}, false},
		{"lte_field_fail", map[string]any{"a": 10, "b": 5}, map[string]any{"a": "numeric|lte_field:b"}, true},

		// gt_date → after
		{"gt_date_pass", map[string]any{"d": "2025-01-02"}, map[string]any{"d": "gt_date:2025-01-01"}, false},
		{"gt_date_fail", map[string]any{"d": "2025-01-01"}, map[string]any{"d": "gt_date:2025-01-02"}, true},

		// lt_date → before
		{"lt_date_pass", map[string]any{"d": "2025-01-01"}, map[string]any{"d": "lt_date:2025-01-02"}, false},
		{"lt_date_fail", map[string]any{"d": "2025-01-02"}, map[string]any{"d": "lt_date:2025-01-01"}, true},

		// gte_date → after_or_equal
		{"gte_date_pass_equal", map[string]any{"d": "2025-01-01"}, map[string]any{"d": "gte_date:2025-01-01"}, false},
		{"gte_date_pass_after", map[string]any{"d": "2025-01-02"}, map[string]any{"d": "gte_date:2025-01-01"}, false},
		{"gte_date_fail", map[string]any{"d": "2024-12-31"}, map[string]any{"d": "gte_date:2025-01-01"}, true},

		// lte_date → before_or_equal
		{"lte_date_pass_equal", map[string]any{"d": "2025-01-01"}, map[string]any{"d": "lte_date:2025-01-01"}, false},
		{"lte_date_pass_before", map[string]any{"d": "2024-12-31"}, map[string]any{"d": "lte_date:2025-01-01"}, false},
		{"lte_date_fail", map[string]any{"d": "2025-01-02"}, map[string]any{"d": "lte_date:2025-01-01"}, true},

		// number → numeric
		{"number_pass_int", map[string]any{"v": 42}, map[string]any{"v": "number"}, false},
		{"number_pass_string", map[string]any{"v": "3.14"}, map[string]any{"v": "number"}, false},
		{"number_fail", map[string]any{"v": "abc"}, map[string]any{"v": "number"}, true},

		// full_url → url
		{"full_url_pass", map[string]any{"v": "https://goravel.dev"}, map[string]any{"v": "full_url"}, false},
		{"full_url_fail", map[string]any{"v": "not-a-url"}, map[string]any{"v": "full_url"}, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := s.makeValidator(tt.data, tt.rules)
			s.Equal(tt.fails, v.Fails())
		})
	}
}
