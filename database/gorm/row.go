package gorm

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"

	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/support/str"
)

type Row struct {
	err   error
	query *Query
	row   map[string]any
}

func (r *Row) Err() error {
	return r.err
}

func (r *Row) Scan(value any) error {
	if r.err != nil {
		return r.err
	}

	msConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			db.ToTimeHookFunc(), db.ToCarbonHookFunc(), db.ToDeletedAtHookFunc(),
		),
		Squash: true,
		Result: value,
		MatchName: func(mapKey, fieldName string) bool {
			return str.Of(mapKey).Studly().String() == fieldName || strings.EqualFold(mapKey, fieldName)
		},
	}

	decoder, err := mapstructure.NewDecoder(msConfig)
	if err != nil {
		return err
	}

	if err := decoder.Decode(r.row); err != nil {
		return err
	}

	for _, item := range r.query.conditions.with {
		// Need to new a query, avoid to clear the conditions
		query := r.query.new(r.query.instance)
		// The new query must be cleared
		query.clearConditions()
		if err := query.Load(value, item.query, item.args...); err != nil {
			return err
		}
	}

	return nil
}
