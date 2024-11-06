package schema

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type IndexDefinition struct {
	command *schema.Command
}

func NewIndexDefinition(command *schema.Command) schema.IndexDefinition {
	return &IndexDefinition{
		command: command,
	}
}

func (f *IndexDefinition) Algorithm(algorithm string) schema.IndexDefinition {
	f.command.Algorithm = algorithm

	return f
}

func (f *IndexDefinition) Name(name string) schema.IndexDefinition {
	f.command.Index = name

	return f
}
