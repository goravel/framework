package validation

import (
	"context"
	"fmt"
	"testing"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

// --- End-to-end validation benchmarks ---

func BenchmarkValidation_Simple(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   "25",
	}
	rules := map[string]any{
		"name":  "required|string|max:255",
		"email": "required|email",
		"age":   "required|integer|min:1|max:150",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = v.Make(ctx, data, rules)
	}
}

func BenchmarkValidation_ManyFields(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	data := make(map[string]any, 50)
	rules := make(map[string]any, 50)
	for i := range 50 {
		key := fmt.Sprintf("field_%d", i)
		data[key] = fmt.Sprintf("value_%d", i)
		rules[key] = "required|string|min:1|max:255"
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = v.Make(ctx, data, rules)
	}
}

func BenchmarkValidation_ComplexRules(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	data := map[string]any{
		"email":                 "test@example.com",
		"website":               "https://example.com",
		"code":                  "ABC-123",
		"start_date":            "2024-01-01",
		"end_date":              "2024-12-31",
		"password":              "secret123",
		"password_confirmation": "secret123",
	}
	rules := map[string]any{
		"email":      "required|email",
		"website":    "required|url",
		"code":       []string{"required", "regex:^[A-Z]{3}-[0-9]{3}$"},
		"start_date": "required|date|before:end_date",
		"end_date":   "required|date|after:start_date",
		"password":   "required|string|min:8|confirmed",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = v.Make(ctx, data, rules)
	}
}

func BenchmarkValidation_Wildcards(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	items := make([]any, 20)
	for i := range 20 {
		items[i] = map[string]any{
			"name":  fmt.Sprintf("Item %d", i),
			"price": fmt.Sprintf("%d.99", i+1),
			"qty":   fmt.Sprintf("%d", i+1),
		}
	}
	data := map[string]any{"items": items}
	rules := map[string]any{
		"items":         "required|array|min:1",
		"items.*.name":  "required|string|max:100",
		"items.*.price": "required|numeric|min:0",
		"items.*.qty":   "required|integer|min:1",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = v.Make(ctx, data, rules)
	}
}

func BenchmarkValidation_AllInvalid(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	data := map[string]any{
		"name":  "",
		"email": "not-an-email",
		"age":   "not-a-number",
		"url":   "not-a-url",
		"date":  "not-a-date",
	}
	rules := map[string]any{
		"name":  "required|string|min:1",
		"email": "required|email",
		"age":   "required|integer|min:1|max:150",
		"url":   "required|url",
		"date":  "required|date",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = v.Make(ctx, data, rules)
	}
}

func BenchmarkValidation_WithFilters(b *testing.B) {
	v := NewValidation()
	ctx := context.Background()
	rules := map[string]any{
		"name":  "required|string|max:255",
		"email": "required|email",
		"bio":   "required|string",
	}
	opts := []contractsvalidation.Option{
		Filters(map[string]any{
			"name":  "trim",
			"email": "trim|lower",
			"bio":   "trim|strip_tags",
		}),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Data is mutated by filters, so recreate it each iteration
		dataCopy := map[string]any{
			"name":  "  John Doe  ",
			"email": "  JOHN@EXAMPLE.COM  ",
			"bio":   "  <b>Hello</b> World  ",
		}
		_, _ = v.Make(ctx, dataCopy, rules, opts...)
	}
}

// --- Component benchmarks ---

func BenchmarkRuleParser_Simple(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ParseRules("required|string|max:255")
	}
}

func BenchmarkRuleParser_Complex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ParseRules("required|string|min:1|max:255|in:a,b,c,d,e")
	}
}

func BenchmarkDataBag_Get_Flat(b *testing.B) {
	bag, _ := NewDataBag(map[string]any{"name": "John", "age": 25})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = bag.Get("name")
	}
}

func BenchmarkDataBag_Get_Nested(b *testing.B) {
	bag, _ := NewDataBag(map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"name": "John",
			},
		},
	})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = bag.Get("user.profile.name")
	}
}

func BenchmarkExpandWildcardFields(b *testing.B) {
	rules := map[string][]ParsedRule{
		"items.*.name":  {{Name: "required"}, {Name: "string"}},
		"items.*.price": {{Name: "required"}, {Name: "numeric"}},
		"items.*.qty":   {{Name: "required"}, {Name: "integer"}},
	}
	keys := make([]string, 0, 100)
	for i := range 20 {
		keys = append(keys, fmt.Sprintf("items.%d", i))
		keys = append(keys, fmt.Sprintf("items.%d.name", i))
		keys = append(keys, fmt.Sprintf("items.%d.price", i))
		keys = append(keys, fmt.Sprintf("items.%d.qty", i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = expandWildcardFields(rules, keys, true)
	}
}
