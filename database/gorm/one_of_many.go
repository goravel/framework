package gorm

import (
	"fmt"
	"strings"

	gormio "gorm.io/gorm"
)

// applyOneOfManyJoin rewrites the inner eager-load query to keep only the row whose value of
// cfg.column equals the per-parent aggregate of cfg.column. The rewrite is an INNER JOIN against
// a grouped subquery; e.g. for MorphOne with cfg.aggregate = "MAX":
//
//	INNER JOIN (
//	    SELECT MAX(created_at) AS aggregate, imageable_id, imageable_type
//	    FROM   images
//	    GROUP BY imageable_id, imageable_type
//	) sub ON images.created_at = sub.aggregate
//	   AND images.imageable_id = sub.imageable_id
//	   AND images.imageable_type = sub.imageable_type
//
// HasOne uses just the foreign-key column in the GROUP BY / JOIN. Mirrors fedaco's
// mixinCanBeOneOfMany at libs/fedaco/src/fedaco/relations/concerns/can-be-one-of-many.ts:81-145.
func (r *Query) applyOneOfManyJoin(inner *gormio.DB, desc *relationDescriptor, cfg *oneOfManyConfig) *gormio.DB {
	if cfg == nil {
		return inner
	}
	switch desc.kind {
	case relKindHasOne, relKindMorphOne:
		// supported
	default:
		// Ignore on unsupported kinds rather than erroring so a generic with-callback that calls
		// LatestOfMany doesn't break unrelated relations downstream. Document elsewhere that
		// OfMany is only effective on HasOne / MorphOne.
		return inner
	}

	if len(desc.references) == 0 {
		return inner
	}
	ref := desc.references[0]
	relatedTable := desc.relatedTable
	const alias = "goravel_one_of_many"

	// Group by the foreign key (and morph type column for MorphOne).
	groupCols := []string{quoteIdent(relatedTable) + "." + quoteIdent(ref.foreignColumn)}
	selectCols := []string{
		fmt.Sprintf("%s(%s.%s) AS aggregate", cfg.aggregate, quoteIdent(relatedTable), quoteIdent(cfg.column)),
		quoteIdent(relatedTable) + "." + quoteIdent(ref.foreignColumn),
	}
	if desc.kind == relKindMorphOne {
		groupCols = append(groupCols, quoteIdent(relatedTable)+"."+quoteIdent(desc.morphTypeColumn))
		selectCols = append(selectCols, quoteIdent(relatedTable)+"."+quoteIdent(desc.morphTypeColumn))
	}

	sub := r.freshSession().
		Table(relatedTable).
		Select(selectCols).
		Group(strings.Join(groupCols, ", "))

	on := []string{
		fmt.Sprintf("%s.%s = %s.aggregate", quoteIdent(relatedTable), quoteIdent(cfg.column), alias),
		fmt.Sprintf("%s.%s = %s.%s", quoteIdent(relatedTable), quoteIdent(ref.foreignColumn), alias, quoteIdent(ref.foreignColumn)),
	}
	if desc.kind == relKindMorphOne {
		on = append(on, fmt.Sprintf("%s.%s = %s.%s",
			quoteIdent(relatedTable), quoteIdent(desc.morphTypeColumn),
			alias, quoteIdent(desc.morphTypeColumn)))
	}

	return inner.Joins(fmt.Sprintf("INNER JOIN (?) %s ON %s", alias, strings.Join(on, " AND ")), sub)
}
