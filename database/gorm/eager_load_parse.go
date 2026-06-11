package gorm

import (
	"strings"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// eagerLoadEntry is the normalised representation of one entry on the eager-load list. It is the
// Go equivalent of fedaco's Record<string, FedacoBuilderCallBack>, but uses a slice in the caller
// to preserve insertion order (Go maps don't).
type eagerLoadEntry struct {
	relation string                        // e.g. "Books" or "Books.Author"
	columns  []string                      // pruned column list parsed from "Books:id,name"; nil = SELECT *
	callback contractsorm.RelationCallback // nil for the synthetic noop entries that _addNestedWiths inserts
}

// parseEagerLoad normalises the variadic args accepted by Query.With into an ordered
// slice of eagerLoadEntry. Mirrors the union of fedaco's _parseWithRelations,
// _addNestedWiths and _createSelectWithConstraint, expressed as Go runtime type-dispatch since Go
// doesn't have TypeScript-style overloads.
//
// Accepted shapes (any of which may also appear inside a single []any):
//   - "Books"
//   - "Books:id,name"                                   (column-pruned)
//   - "Books.Author"                                    (nested; auto-fills "Books" as a noop entry)
//   - "Books", callback                                 (string + callback as the only two args)
//   - "Books", "Roles", "Address"                       (multiple strings)
//   - map[string]contractsorm.RelationCallback{...}     (relation -> callback)
//   - []string{"Books", "Roles"}
//   - []any{"Books", map[string]contractsorm.RelationCallback{"Roles": cb}}
func parseEagerLoad(args []any) ([]eagerLoadEntry, error) {
	// Special-case the (string, callback) two-arg form so q.With("Books", cb) binds the
	// callback to the string rather than treating cb as a freestanding entry.
	if len(args) == 2 {
		if name, ok := args[0].(string); ok {
			if cb, ok := toRelationCallback(args[1]); ok {
				return appendEagerLoadEntry(nil, name, cb)
			}
		}
	}

	var out []eagerLoadEntry
	for _, arg := range args {
		var err error
		out, err = mergeEagerLoadArg(out, arg)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// mergeEagerLoadArg dispatches a single arg into out, handling all accepted shapes recursively.
func mergeEagerLoadArg(out []eagerLoadEntry, arg any) ([]eagerLoadEntry, error) {
	switch v := arg.(type) {
	case nil:
		return out, nil
	case string:
		return appendEagerLoadEntry(out, v, nil)
	case []string:
		var err error
		for _, s := range v {
			out, err = appendEagerLoadEntry(out, s, nil)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case []any:
		var err error
		for _, item := range v {
			out, err = mergeEagerLoadArg(out, item)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case map[string]contractsorm.RelationCallback:
		return appendEagerLoadMap(out, v)
	case map[string]func(contractsorm.Query) contractsorm.Query:
		converted := make(map[string]contractsorm.RelationCallback, len(v))
		for k, fn := range v {
			converted[k] = contractsorm.RelationCallback(fn)
		}
		return appendEagerLoadMap(out, converted)
	default:
		return nil, errors.OrmEagerLoadInvalidArgument.Args(v)
	}
}

func appendEagerLoadMap(out []eagerLoadEntry, m map[string]contractsorm.RelationCallback) ([]eagerLoadEntry, error) {
	var err error
	for name, cb := range m {
		out, err = appendEagerLoadEntry(out, name, cb)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// appendEagerLoadEntry adds one (relation, callback) pair to out, applying _addNestedWiths and
// _createSelectWithConstraint semantics:
//   - "A.B.C" inserts noop entries for each missing prefix (A, A.B), then a real entry for A.B.C
//   - "Books:id,name" splits into name="Books" + columns=[id, name]
//   - duplicate relations: the later write replaces the earlier (last-wins) while preserving the
//     position of the earlier entry, matching fedaco's overwrite-in-place behaviour for
//     Record<string, ...>
func appendEagerLoadEntry(out []eagerLoadEntry, raw string, cb contractsorm.RelationCallback) ([]eagerLoadEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.OrmEagerLoadEmptyRelation
	}

	name, columns := splitRelationSelect(raw)
	if name == "" {
		return nil, errors.OrmEagerLoadEmptyRelation
	}

	// Walk dot-segments and ensure every prefix has an entry. Only the leaf carries the real
	// callback / columns; intermediate prefixes get a synthetic placeholder.
	segments := strings.Split(name, ".")
	progress := ""
	for i, seg := range segments {
		if seg == "" {
			return nil, errors.OrmEagerLoadEmptyRelation
		}
		if i == 0 {
			progress = seg
		} else {
			progress = progress + "." + seg
		}
		isLeaf := i == len(segments)-1
		entry := eagerLoadEntry{relation: progress}
		if isLeaf {
			entry.columns = columns
			entry.callback = cb
		}
		out = upsertEagerLoadEntry(out, entry, isLeaf)
	}
	return out, nil
}

// upsertEagerLoadEntry inserts entry into out, or — when an entry with the same relation already
// exists — overwrites it in place. The isLeaf flag prevents synthetic prefix placeholders from
// clobbering an existing real entry: if "Books" was already added with a callback, walking
// through "Books.Author" should not erase Books's callback when re-touching the prefix.
func upsertEagerLoadEntry(out []eagerLoadEntry, entry eagerLoadEntry, isLeaf bool) []eagerLoadEntry {
	for i, existing := range out {
		if existing.relation == entry.relation {
			if isLeaf {
				out[i] = entry
			}
			return out
		}
	}
	return append(out, entry)
}

// splitRelationSelect splits "Books:id,name" into ("Books", ["id", "name"]). A bare relation name
// without a colon yields (name, nil). Whitespace inside the column list is trimmed.
func splitRelationSelect(raw string) (string, []string) {
	prefix, rest, ok := strings.Cut(raw, ":")
	if !ok {
		return raw, nil
	}
	name := strings.TrimSpace(prefix)
	if rest == "" {
		return name, nil
	}
	parts := strings.Split(rest, ",")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cols = append(cols, p)
		}
	}
	if len(cols) == 0 {
		return name, nil
	}
	return name, cols
}

// toRelationCallback accepts the two callable shapes a user is likely to pass and converts to a
// canonical contractsorm.RelationCallback. Returns ok=false for anything else.
func toRelationCallback(v any) (contractsorm.RelationCallback, bool) {
	switch fn := v.(type) {
	case nil:
		return nil, true
	case contractsorm.RelationCallback:
		return fn, true
	case func(contractsorm.Query) contractsorm.Query:
		return contractsorm.RelationCallback(fn), true
	}
	return nil, false
}

// directNestedEntries returns the entries from list whose relation is strictly nested under
// parent (i.e. starts with "parent."), with the "parent." prefix stripped. Used when recursing
// into a child query: "Books.Author" under parent "Books" becomes "Author".
func directNestedEntries(list []eagerLoadEntry, parent string) []eagerLoadEntry {
	prefix := parent + "."
	var out []eagerLoadEntry
	for _, e := range list {
		if strings.HasPrefix(e.relation, prefix) {
			child := e
			child.relation = strings.TrimPrefix(e.relation, prefix)
			out = append(out, child)
		}
	}
	return out
}
