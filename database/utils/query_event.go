package utils

import (
	"context"
	"sync"

	"gorm.io/gorm"

	contractsdb "github.com/goravel/framework/contracts/database/db"
)

// queryListeners holds the listeners registered through Listen. It is a
// package-level registry so that a listener registered on one connection (or on
// the DB facade) observes queries executed on every connection, matching the
// global nature of Laravel's DB::listen.
var queryListeners = &listenerRegistry{}

type listenerRegistry struct {
	mu        sync.RWMutex
	listeners []func(event *contractsdb.QueryExecuted) error
}

func (r *listenerRegistry) add(listener func(event *contractsdb.QueryExecuted) error) {
	if listener == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners = append(r.listeners, listener)
}

func (r *listenerRegistry) has() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.listeners) > 0
}

func (r *listenerRegistry) dispatch(event *contractsdb.QueryExecuted) {
	r.mu.RLock()
	listeners := r.listeners
	r.mu.RUnlock()

	for _, listener := range listeners {
		r.call(listener, event)
	}
}

func (r *listenerRegistry) call(listener func(event *contractsdb.QueryExecuted) error, event *contractsdb.QueryExecuted) {
	// A panicking listener must never break the query flow, so the panic is
	// recovered and swallowed here.
	defer func() {
		_ = recover()
	}()

	_ = listener(event)
}

// ListenQuery registers a listener that is called for every SQL query executed
// by the database, mirroring Laravel's DB::listen.
func ListenQuery(listener func(event *contractsdb.QueryExecuted) error) {
	queryListeners.add(listener)
}

// ClearQueryListeners removes all registered query event listeners. It exists
// for tests that register listeners through the global registry.
func ClearQueryListeners() {
	queryListeners.mu.Lock()
	defer queryListeners.mu.Unlock()
	queryListeners.listeners = nil
}

// HasQueryListeners reports whether any query event listener is registered.
// Hot paths use it to skip building events nobody would receive.
func HasQueryListeners() bool {
	return queryListeners.has()
}

// DispatchQueryEvent delivers the event to every registered listener.
func DispatchQueryEvent(event *contractsdb.QueryExecuted) {
	queryListeners.dispatch(event)
}

type queryBindingsKey struct{}

type queryBindings struct {
	sql  string
	args []any
}

// WithQueryBindings carries the placeholder SQL and its bound arguments
// alongside the context so that the query-event dispatcher can report them.
func WithQueryBindings(ctx context.Context, sql string, args []any) context.Context {
	return context.WithValue(ctx, queryBindingsKey{}, queryBindings{sql: sql, args: args})
}

// QueryBindingsFromContext returns the bindings carried on the context, if any.
func QueryBindingsFromContext(ctx context.Context) (string, []any, bool) {
	bindings, ok := ctx.Value(queryBindingsKey{}).(queryBindings)
	if !ok {
		return "", nil, false
	}

	return bindings.sql, bindings.args, true
}

const QueryEventPluginName = "goravel:query_event"

// QueryEventPlugin is a gorm plugin that carries the built SQL and its
// variables on the statement context right before gorm traces the query, so
// the logger can emit a QueryExecuted event with Sql and Bindings populated.
type QueryEventPlugin struct{}

func NewQueryEventPlugin() *QueryEventPlugin {
	return &QueryEventPlugin{}
}

func (r *QueryEventPlugin) Name() string {
	return QueryEventPluginName
}

func (r *QueryEventPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()

	if err := cb.Create().After("gorm:create").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}
	if err := cb.Query().After("gorm:query").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}
	if err := cb.Update().After("gorm:update").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}
	if err := cb.Delete().After("gorm:delete").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}
	if err := cb.Row().After("gorm:row").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}
	if err := cb.Raw().After("gorm:raw").Register(QueryEventPluginName+":after", r.after); err != nil {
		return err
	}

	return nil
}

func (r *QueryEventPlugin) after(tx *gorm.DB) {
	if !HasQueryListeners() || tx.Statement == nil || tx.Statement.SQL.Len() == 0 {
		return
	}

	sql := tx.Statement.SQL.String()
	vars := tx.Statement.Vars

	tx.Statement.Context = WithQueryBindings(tx.Statement.Context, sql, append([]any{}, vars...))
}
