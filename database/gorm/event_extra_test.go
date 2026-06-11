package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"
)

// TestEvent_Context_NilCtx verifies Context returns the underlying query's context (nil when
// unset).
func TestEvent_Context_NilCtx(t *testing.T) {
	q := &Query{instance: &gormio.DB{Statement: &gormio.Statement{Selects: []string{}, Omits: []string{}}}}
	e := NewEvent(q, &testEventModel, nil)
	assert.Nil(t, e.Context())
}

// TestEvent_Context_ReturnsCtx verifies Context returns the configured context.
func TestEvent_Context_ReturnsCtx(t *testing.T) {
	type ctxKey int
	const k ctxKey = 1
	ctx := context.WithValue(context.Background(), k, "value")
	q := &Query{
		ctx:      ctx,
		instance: &gormio.DB{Statement: &gormio.Statement{Selects: []string{}, Omits: []string{}}},
	}
	e := NewEvent(q, &testEventModel, nil)
	assert.Equal(t, "value", e.Context().Value(k))
}

// TestEvent_IsClean_InverseOfIsDirty verifies IsClean returns !IsDirty.
func TestEvent_IsClean_InverseOfIsDirty(t *testing.T) {
	// When dest is empty/nil, IsDirty returns false, so IsClean returns true.
	q := &Query{instance: &gormio.DB{Statement: &gormio.Statement{Selects: []string{}, Omits: []string{}}}}
	e := NewEvent(q, &testEventModel, nil)
	assert.True(t, e.IsClean())
	assert.False(t, e.IsDirty())
}
