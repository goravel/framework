# Many-to-Many Relations (BelongsToMany family)

The BelongsToMany family covers three relation kinds — all share a pivot table:

| Kind | Use case |
|---|---|
| `Many2Many` | Plain m:n between two models (e.g. User ↔ Role through `user_roles`) |
| `MorphToMany` | Polymorphic m:n where the parent side is morphable (e.g. Post → Tag, Video → Tag through `taggables`) |
| `MorphedByMany` | Inverse of `MorphToMany` (e.g. Tag → Post, Tag → Video) |

All three accept the same set of pivot configuration fields described below.

---

## Basic declaration

```go
type User struct {
    ID    uint
    Roles []*Role `gorm:"-"`
}

func (User) Relations() map[string]orm.Relation {
    return map[string]orm.Relation{
        "Roles": orm.Many2Many{Related: &Role{}, Table: "user_roles"},
    }
}
```

`Table` is optional — the framework defaults to the alphabetically-sorted singular pair (e.g. `role_user`).

## Pivot model (custom Pivot struct)

To surface pivot-table columns on eager-loaded results, declare a Go struct for the pivot row and add a field of that type to the related model:

```go
type RoleUserPivot struct {
    UserID    uint      `gorm:"column:user_id"`
    RoleID    uint      `gorm:"column:role_id"`
    Active    bool      `gorm:"column:active"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Role struct {
    ID    uint
    Name  string
    Pivot RoleUserPivot `gorm:"-"`   // ← framework hydrates this on eager load
}
```

No additional config required — the framework reflects the `Pivot` field type, parses its GORM schema, and uses the resulting `db_name` columns as the pivot SELECT list.

### Custom pivot field name (`PivotField`)

If a single related model participates in multiple m:n relations with different pivot schemas, use a different field name per relation:

```go
type Role struct {
    ID         uint
    Name       string
    UserPivot  RoleUserPivot  `gorm:"-"`
    GroupPivot RoleGroupPivot `gorm:"-"`
}

// On User
"Roles": orm.Many2Many{Related: &Role{}, PivotField: "UserPivot"}

// On Group
"Roles": orm.Many2Many{Related: &Role{}, PivotField: "GroupPivot"}
```

`PivotField` defaults to `"Pivot"` when omitted.

## Auto-stamping pivot timestamps

The framework auto-fills `created_at` (on INSERT) and `updated_at` (on INSERT/UPDATE) using the following priority order:

1. **Explicit GORM tag** on a Pivot struct field — works for any field name:
   ```go
   type RoleUserPivot struct {
       Stamped time.Time `gorm:"autoCreateTime"`
       Edited  time.Time `gorm:"autoUpdateTime"`
   }
   ```

2. **GORM convention** — fields named `CreatedAt` / `UpdatedAt` of type `time.Time`:
   ```go
   type RoleUserPivot struct {
       CreatedAt time.Time
       UpdatedAt time.Time
   }
   ```

3. **Relation-level fallback** — set `PivotTimestamps: true` when no Pivot struct is declared (or it has no timestamp fields) but the underlying table still has `created_at` / `updated_at` you want auto-filled:
   ```go
   "Roles": orm.Many2Many{Related: &Role{}, PivotTimestamps: true}
   ```

### Customising column names

Use the GORM `column` tag on the Pivot struct field — there is no relation-level override:

```go
type RoleUserPivot struct {
    Stamped time.Time `gorm:"autoCreateTime;column:made_on"`
    Edited  time.Time `gorm:"autoUpdateTime;column:edited_at"`
}
```

### Partial timestamps

Declaring only `CreatedAt` (or only `UpdatedAt`) is fine — the framework only auto-stamps what you declare:

```go
type RoleUserPivot struct {
    UserID    uint
    RoleID    uint
    CreatedAt time.Time   // only created_at gets stamped; updated_at is left untouched
}
```

## Filtering pivot operations (`OnPivotQuery`)

Sync / Attach / Detach / Toggle / UpdateExistingPivot can be scoped to a subset of pivot rows via `OnPivotQuery` — equivalent to fedaco's `wherePivot` / `wherePivotIn` / `wherePivotNull`.

```go
"ActiveRoles": orm.Many2Many{
    Related: &Role{},
    Table:   "user_roles",
    OnPivotQuery: func(q orm.PivotQuery) orm.PivotQuery {
        return q.Where("active", 1).WhereNull("deleted_at")
    },
}
```

The callback applies to **SELECT / UPDATE / DELETE** on the pivot table:

- `Sync` reads "current" rows through the filter — so it only detaches rows matching the filter.
- `Detach` only deletes rows matching the filter.
- `UpdateExistingPivot` only updates rows matching the filter.
- `Attach`'s duplicate-detection SELECT also goes through the filter (so a row not matching the filter is treated as "not attached").

INSERT rows from `Attach` / `AttachWithPivot` are **not** auto-injected with these conditions — pass equality columns through the `attrs` map:

```go
orm.Attach(&user, "ActiveRoles", []any{1, 2, 3})
// Will INSERT (user_id, role_id) only — without active=1 unless you add it explicitly:
orm.AttachWithPivot(&user, "ActiveRoles", map[any]map[string]any{
    1: {"active": 1},
    2: {"active": 1},
})
```

### `PivotQuery` interface

```go
type PivotQuery interface {
    Where(column string, args ...any) PivotQuery
    WhereIn(column string, values []any) PivotQuery
    WhereNotIn(column string, values []any) PivotQuery
    WhereNull(column string) PivotQuery
    WhereNotNull(column string) PivotQuery
}
```

## Bumping parent timestamps (`Touches`)

Pivot writes don't affect the parent row's `updated_at` by default. Set `Touches: true` to make Sync / Attach / Detach / Toggle / UpdateExistingPivot bump the parent's `updated_at` after the pivot operation succeeds (and only when pivot rows actually changed):

```go
type Post struct {
    ID        uint
    Title     string
    UpdatedAt time.Time
    Tags      []*Tag `gorm:"-"`
}

func (Post) Relations() map[string]orm.Relation {
    return map[string]orm.Relation{
        "Tags": orm.MorphToMany{Related: &Tag{}, Name: "taggable", Touches: true},
    }
}

orm.Sync(&post, "Tags", []any{1, 2, 3})
// → UPDATE posts SET updated_at = NOW() WHERE id = ?
```

Silently no-ops when:
- The relation isn't `Touches: true`.
- The parent's schema has no `updated_at` field.
- The pivot operation didn't actually attach/detach/update any rows (e.g. `Sync` with the same id list as already attached).

## Pivot operation API

All pivot operations live on `Orm` (and on the underlying `Query`):

| Method | Effect |
|---|---|
| `Attach(parent, rel, ids)` | INSERT pivot rows; skips ids already attached |
| `AttachWithPivot(parent, rel, idsWithAttrs)` | Attach with per-id pivot column values |
| `Detach(parent, rel, ids...)` | DELETE pivot rows; with no ids, detaches all |
| `Sync(parent, rel, ids)` | INSERT missing + DELETE extra; idempotent |
| `SyncWithPivot(parent, rel, idsWithAttrs)` | Sync with per-id pivot values; UPDATEs existing rows when attrs non-empty |
| `SyncWithPivotValues(parent, rel, ids, sharedAttrs)` | Convenience: all ids share the same pivot values |
| `SyncWithoutDetaching(parent, rel, ids)` | INSERT missing only — no detach |
| `SyncWithoutDetachingWithPivot(parent, rel, idsWithAttrs)` | Same, with attrs |
| `Toggle(parent, rel, ids)` | Attach missing, detach existing |
| `ToggleWithPivot(parent, rel, idsWithAttrs)` | Toggle with attrs on the attach side |
| `UpdateExistingPivot(parent, rel, id, attrs)` | UPDATE one already-attached pivot row |

### `SyncResult`

`Sync*` and `Toggle*` return `*db.SyncResult`:

```go
type SyncResult struct {
    Attached []any
    Detached []any
    Updated  []any  // populated by SyncWithPivot* when an existing row's attrs changed
}
```

Element type matches the related model's primary-key Go type (e.g. `uint`, `int64`, `string`) — caller-supplied ids are normalised regardless of input type.

## Complete example

```go
import (
    "time"

    "github.com/goravel/framework/contracts/database/orm"
)

type RoleUserPivot struct {
    UserID    uint      `gorm:"column:user_id"`
    RoleID    uint      `gorm:"column:role_id"`
    Active    bool      `gorm:"column:active"`
    Priority  int       `gorm:"column:priority"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Role struct {
    ID    uint
    Name  string
    Pivot RoleUserPivot `gorm:"-"`
}

type User struct {
    ID        uint
    Name      string
    UpdatedAt time.Time
    Roles     []*Role `gorm:"-"`
}

func (User) Relations() map[string]orm.Relation {
    return map[string]orm.Relation{
        "Roles": orm.Many2Many{
            Related: &Role{},
            Table:   "user_roles",
            OnPivotQuery: func(q orm.PivotQuery) orm.PivotQuery {
                return q.Where("active", 1)
            },
            Touches: true,
        },
    }
}

// Usage:
user := User{ID: 7}
orm := facades.Orm()

// Attach roles 1 and 2 with pivot data.
orm.AttachWithPivot(&user, "Roles", map[any]map[string]any{
    uint(1): {"active": 1, "priority": 10},
    uint(2): {"active": 1, "priority": 5},
})
// Pivot rows inserted with active=1, priority=N, created_at=NOW(), updated_at=NOW().
// User.UpdatedAt also bumped (Touches: true).

// Sync to roles {1, 3}: detaches 2, attaches 3, leaves 1 untouched.
result, _ := orm.Sync(&user, "Roles", []any{uint(1), uint(3)})
// result.Attached = [3], result.Detached = [2], result.Updated = []

// Eager load roles with pivot data.
var u User
orm.Query().With("Roles").First(&u, uint(7))
// u.Roles[0].Pivot.Active == true, u.Roles[0].Pivot.Priority == 10, etc.
```
