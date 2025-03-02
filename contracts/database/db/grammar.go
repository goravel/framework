package db

import (
	sq "github.com/Masterminds/squirrel"
)

type CompileLimitGrammar interface {
	CompileLimit(builder sq.SelectBuilder, limit uint64) sq.SelectBuilder
}

type CompileOffsetGrammar interface {
	CompileOffset(builder sq.SelectBuilder, offset uint64) sq.SelectBuilder
}

type CompileOrderByGrammar interface {
	CompileOrderBy(builder sq.SelectBuilder, raw []string) sq.SelectBuilder
}
