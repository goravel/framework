package processors

import (
	"strings"

	"github.com/spf13/cast"

	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

type Postgres struct {
}

func NewPostgres() *Postgres {
	return &Postgres{}
}

func (r *Postgres) ProcessColumns(columns []schemacontract.Column) []schemacontract.Column {
	for i, column := range columns {
		columns[i].AutoIncrement = column.Default != "" && strings.HasPrefix(cast.ToString(column.Default), "nextval(")
	}

	return columns
}

func (r *Postgres) ProcessIndexes(indexes []schemacontract.Index) []schemacontract.Index {
	for i, index := range indexes {
		indexes[i].Name = strings.ToLower(index.Name)
		indexes[i].Type = strings.ToLower(index.Type)
	}

	return indexes
}
