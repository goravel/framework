package gorm

import (
	dbcontract "github.com/goravel/framework/contracts/database/db"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// relationWriter binds (parent, name) to a Query session and forwards each write to the
// matching *Relation-suffixed method on *Query. It implements contractsorm.RelationWriter so
// callers can use a single chained entry — Query.Relation(parent, name) — for all writes.
type relationWriter struct {
	q      *Query
	parent any
	name   string
}

// Relation returns a RelationWriter bound to the given parent and relation name. The returned
// builder forwards all write operations to the receiver's session, so calls inside a Transaction
// callback honor the transaction.
func (r *Query) Relation(parent any, name string) contractsorm.RelationWriter {
	return &relationWriter{q: r, parent: parent, name: name}
}

func (w *relationWriter) Save(child any) error {
	return w.q.SaveRelation(w.parent, w.name, child)
}

func (w *relationWriter) SaveMany(children any) error {
	return w.q.SaveManyRelation(w.parent, w.name, children)
}

func (w *relationWriter) SaveWithPivot(child any, attrs map[string]any) error {
	return w.q.SaveRelationWithPivot(w.parent, w.name, child, attrs)
}

func (w *relationWriter) SaveManyWithPivot(children any, attrsPerChild map[any]map[string]any) error {
	return w.q.SaveManyRelationWithPivot(w.parent, w.name, children, attrsPerChild)
}

func (w *relationWriter) Create(dest any) error {
	return w.q.CreateRelation(w.parent, w.name, dest)
}

func (w *relationWriter) CreateMany(dests any) error {
	return w.q.CreateManyRelation(w.parent, w.name, dests)
}

func (w *relationWriter) FindOrNew(id any, dest any) error {
	return w.q.FindOrNewRelation(w.parent, w.name, id, dest)
}

func (w *relationWriter) FirstOrNew(attrs, values map[string]any, dest any) error {
	return w.q.FirstOrNewRelation(w.parent, w.name, attrs, values, dest)
}

func (w *relationWriter) FirstOrCreate(attrs, values map[string]any, dest any) error {
	return w.q.FirstOrCreateRelation(w.parent, w.name, attrs, values, dest)
}

func (w *relationWriter) UpdateOrCreate(attrs, values map[string]any, dest any) error {
	return w.q.UpdateOrCreateRelation(w.parent, w.name, attrs, values, dest)
}

func (w *relationWriter) Associate(owner any) error {
	return w.q.AssociateRelation(w.parent, w.name, owner)
}

func (w *relationWriter) Dissociate() error {
	return w.q.DissociateRelation(w.parent, w.name)
}

func (w *relationWriter) Attach(ids []any) error {
	return w.q.AttachRelation(w.parent, w.name, ids)
}

func (w *relationWriter) AttachWithPivot(idsWithAttrs map[any]map[string]any) error {
	return w.q.AttachWithPivotRelation(w.parent, w.name, idsWithAttrs)
}

func (w *relationWriter) Detach(ids ...any) (int64, error) {
	return w.q.DetachRelation(w.parent, w.name, ids)
}

func (w *relationWriter) Sync(ids []any) (*dbcontract.SyncResult, error) {
	return w.q.SyncRelation(w.parent, w.name, ids)
}

func (w *relationWriter) SyncWithPivot(idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return w.q.SyncRelationWithPivot(w.parent, w.name, idsWithAttrs)
}

func (w *relationWriter) SyncWithPivotValues(ids []any, pivotValues map[string]any) (*dbcontract.SyncResult, error) {
	return w.q.SyncRelationWithPivotValues(w.parent, w.name, ids, pivotValues)
}

func (w *relationWriter) SyncWithoutDetaching(ids []any) (*dbcontract.SyncResult, error) {
	return w.q.SyncWithoutDetachingRelation(w.parent, w.name, ids)
}

func (w *relationWriter) SyncWithoutDetachingWithPivot(idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return w.q.SyncWithoutDetachingRelationWithPivot(w.parent, w.name, idsWithAttrs)
}

func (w *relationWriter) Toggle(ids []any) (*dbcontract.SyncResult, error) {
	return w.q.ToggleRelation(w.parent, w.name, ids)
}

func (w *relationWriter) ToggleWithPivot(idsWithAttrs map[any]map[string]any) (*dbcontract.SyncResult, error) {
	return w.q.ToggleRelationWithPivot(w.parent, w.name, idsWithAttrs)
}

func (w *relationWriter) UpdateExistingPivot(id any, attrs map[string]any) (int64, error) {
	return w.q.UpdateExistingPivotRelation(w.parent, w.name, id, attrs)
}
