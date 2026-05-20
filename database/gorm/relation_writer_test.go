package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// TestRelation_ReturnsWriter verifies Query.Relation returns a writer bound to the parent/name.
func TestRelation_ReturnsWriter(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	parent := &relUser{ID: 1}

	w := q.Relation(parent, "Books")
	assert.NotNil(t, w)

	rw, ok := w.(*relationWriter)
	assert.True(t, ok)
	assert.Same(t, q, rw.q)
	assert.Same(t, parent, rw.parent)
	assert.Equal(t, "Books", rw.name)
}

// TestRelation_ImplementsContract verifies relationWriter satisfies the RelationWriter contract.
func TestRelation_ImplementsContract(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_ = contractsorm.RelationWriter(q.Relation(&relUser{}, "Books"))
}

// TestRelationWriter_Save_Delegates verifies Save forwards to SaveRelation.
func TestRelationWriter_Save_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.Save(&relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SaveMany_Delegates verifies SaveMany forwards to SaveManyRelation.
func TestRelationWriter_SaveMany_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.SaveMany([]*relBook{{}})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SaveWithPivot_Delegates verifies SaveWithPivot forwards to SaveRelationWithPivot.
func TestRelationWriter_SaveWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	err := w.SaveWithPivot(&relRole{}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SaveManyWithPivot_Delegates verifies SaveManyWithPivot forwards correctly.
func TestRelationWriter_SaveManyWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	err := w.SaveManyWithPivot([]*relRole{{}}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Create_Delegates verifies Create forwards to CreateRelation.
func TestRelationWriter_Create_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.Create(&relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_CreateMany_Delegates verifies CreateMany forwards to CreateManyRelation.
func TestRelationWriter_CreateMany_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.CreateMany([]*relBook{{}})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_FindOrNew_Delegates verifies FindOrNew forwards to FindOrNewRelation.
func TestRelationWriter_FindOrNew_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.FindOrNew(1, &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_FirstOrNew_Delegates verifies FirstOrNew forwards to FirstOrNewRelation.
func TestRelationWriter_FirstOrNew_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.FirstOrNew(nil, nil, &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_FirstOrCreate_Delegates verifies FirstOrCreate forwards to FirstOrCreateRelation.
func TestRelationWriter_FirstOrCreate_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.FirstOrCreate(nil, nil, &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_UpdateOrCreate_Delegates verifies UpdateOrCreate forwards correctly.
func TestRelationWriter_UpdateOrCreate_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Books") // pass by value to trigger error path

	err := w.UpdateOrCreate(nil, nil, &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Associate_Delegates verifies Associate forwards to AssociateRelation.
func TestRelationWriter_Associate_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	w := q.Relation(relBook{}, "Author") // pass by value to trigger error path

	err := w.Associate(&relUser{ID: 1})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Dissociate_Delegates verifies Dissociate forwards to DissociateRelation.
func TestRelationWriter_Dissociate_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	w := q.Relation(relBook{}, "Author") // pass by value to trigger error path

	err := w.Dissociate()
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Attach_Delegates verifies Attach forwards to AttachRelation.
func TestRelationWriter_Attach_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	err := w.Attach([]any{1, 2})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_AttachWithPivot_Delegates verifies AttachWithPivot forwards correctly.
func TestRelationWriter_AttachWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	err := w.AttachWithPivot(map[any]map[string]any{1: nil})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Detach_Delegates verifies Detach forwards to DetachRelation.
func TestRelationWriter_Detach_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.Detach(1, 2)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Sync_Delegates verifies Sync forwards to SyncRelation.
func TestRelationWriter_Sync_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.Sync([]any{1, 2})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SyncWithPivot_Delegates verifies SyncWithPivot forwards correctly.
func TestRelationWriter_SyncWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.SyncWithPivot(map[any]map[string]any{1: nil})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SyncWithPivotValues_Delegates verifies SyncWithPivotValues forwards correctly.
func TestRelationWriter_SyncWithPivotValues_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.SyncWithPivotValues([]any{1, 2}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SyncWithoutDetaching_Delegates verifies SyncWithoutDetaching forwards.
func TestRelationWriter_SyncWithoutDetaching_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.SyncWithoutDetaching([]any{1, 2})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_SyncWithoutDetachingWithPivot_Delegates verifies the method forwards.
func TestRelationWriter_SyncWithoutDetachingWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.SyncWithoutDetachingWithPivot(map[any]map[string]any{1: nil})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_Toggle_Delegates verifies Toggle forwards to ToggleRelation.
func TestRelationWriter_Toggle_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.Toggle([]any{1, 2})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_ToggleWithPivot_Delegates verifies ToggleWithPivot forwards correctly.
func TestRelationWriter_ToggleWithPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.ToggleWithPivot(map[any]map[string]any{1: nil})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// TestRelationWriter_UpdateExistingPivot_Delegates verifies UpdateExistingPivot forwards.
func TestRelationWriter_UpdateExistingPivot_Delegates(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	w := q.Relation(relUser{ID: 1}, "Roles") // pass by value to trigger error path

	_, err := w.UpdateExistingPivot(1, map[string]any{"k": "v"})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}
