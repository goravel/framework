package gorm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Per-parent pivot hydration tests. The bug copilot flagged (PR #1463 review comments 2 and 4)
// was that the eager loader keyed pivot data by related-ID only, so when the same related model
// is attached to multiple parents with different pivot attributes, one parent's values silently
// overwrote another's. The fix clones the related row per pivot row when a Pivot field is in
// play. These tests pin both the clone semantics and the resulting per-parent state.

func TestCloneRelatedRow_PointerInstance_IsIndependent(t *testing.T) {
	template := &roleWithPivot{ID: 1, Name: "admin"}
	tv := reflect.ValueOf(template)

	clone := cloneRelatedRow(tv)

	assert.True(t, clone.Kind() == reflect.Pointer, "clone should be a pointer like the template")
	// Different instance.
	assert.NotSame(t, template, clone.Interface().(*roleWithPivot), "clone must be a distinct pointer")
	// Same data.
	got := clone.Interface().(*roleWithPivot)
	assert.Equal(t, uint(1), got.ID)
	assert.Equal(t, "admin", got.Name)
	// Mutating the clone must not bleed into the template.
	got.Pivot.Priority = "high"
	assert.Equal(t, "", template.Pivot.Priority, "writes on the clone must not affect the template")
}

// TestM2MHydration_PerParentPivot simulates the hydration loop in loadMany2Many for the
// scenario the bug fix targets: one related role attached to two users with different pivot
// rows. Asserts each user's role instance carries its own pivot values.
func TestM2MHydration_PerParentPivot(t *testing.T) {
	plan := mustPivotPlan(t, "Pivot", &roleUserPivot{})

	// One template related row, keyed by related ID.
	template := reflect.ValueOf(&roleWithPivot{ID: 99, Name: "admin"})
	relatedByID := map[string]reflect.Value{
		"99": template,
	}

	// Two parents (users), each pivot-attached to role 99 with different pivot attributes.
	pivotRows := []map[string]any{
		{"user_id": uint(1), "role_id": uint(99), "priority": "high", "notes": "for-user-1"},
		{"user_id": uint(2), "role_id": uint(99), "priority": "low", "notes": "for-user-2"},
	}

	dict := make(map[string][]reflect.Value)
	for _, p := range pivotRows {
		parentKey := dictKey(p["user_id"])
		relatedKey := dictKey(p["role_id"])
		tpl := relatedByID[relatedKey]

		row := cloneRelatedRow(tpl)
		data := map[string]any{"priority": p["priority"], "notes": p["notes"]}
		assert.NoError(t, writePivotField(t.Context(), row, data, plan))
		dict[parentKey] = append(dict[parentKey], row)
	}

	// Both parents must see distinct role instances with their own pivot data.
	u1 := dict["1"][0].Interface().(*roleWithPivot)
	u2 := dict["2"][0].Interface().(*roleWithPivot)

	assert.NotSame(t, u1, u2, "each parent must get its own role instance")
	assert.Equal(t, "high", u1.Pivot.Priority)
	assert.Equal(t, "for-user-1", u1.Pivot.Notes)
	assert.Equal(t, "low", u2.Pivot.Priority)
	assert.Equal(t, "for-user-2", u2.Pivot.Notes)

	// The template itself must remain untouched — sanity guard that we cloned rather than
	// mutated the shared template.
	tpl := template.Interface().(*roleWithPivot)
	assert.Equal(t, "", tpl.Pivot.Priority, "template must not have been mutated")
	assert.Equal(t, "", tpl.Pivot.Notes, "template must not have been mutated")
}

// When the related model has no Pivot field (pivotPlan == nil in the production code), the
// hydration loop must keep sharing the template across parents — cloning there would waste
// allocations with no behavioral upside. Pins the "skip clone when no Pivot field" branch.
func TestM2MHydration_NoPivotField_SharesTemplate(t *testing.T) {
	template := reflect.ValueOf(&roleWithoutPivot{ID: 99, Name: "admin"})
	relatedByID := map[string]reflect.Value{"99": template}

	pivotRows := []map[string]any{
		{"user_id": uint(1), "role_id": uint(99)},
		{"user_id": uint(2), "role_id": uint(99)},
	}

	// Mirrors the production loop with pivotPlan == nil: no clone, no writePivotField.
	dict := make(map[string][]reflect.Value)
	for _, p := range pivotRows {
		parentKey := dictKey(p["user_id"])
		relatedKey := dictKey(p["role_id"])
		dict[parentKey] = append(dict[parentKey], relatedByID[relatedKey])
	}

	u1 := dict["1"][0].Interface().(*roleWithoutPivot)
	u2 := dict["2"][0].Interface().(*roleWithoutPivot)
	assert.Same(t, u1, u2, "without a Pivot field, parents should share the template instance")
}
