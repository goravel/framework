package migration

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type Creator struct {
}

func NewCreator() *Creator {
	return &Creator{}
}

// GetStub Get the migration stub file.
func (r *Creator) GetStub(table string, create bool) string {
	if table == "" {
		return Stubs{}.Empty()
	}

	if create {
		return Stubs{}.Create()
	}

	return Stubs{}.Update()
}

type StubData struct {
	Package      string
	StructName   string
	Signature    string
	Table        string
	SchemaFields []string
}

// PopulateStub Populate the place-holders in the migration stub.
func (r *Creator) PopulateStub(stub, signature, table string) string {
	data := StubData{
		Package:    "migrations",
		StructName: str.Of(signature).Prepend("m_").Studly().String(),
		Signature:  signature,
		Table:      table,
	}

	tmpl, err := template.New("stub").Parse(stub)
	if err != nil {
		return stub
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return stub
	}

	return buf.String()
}

// GetPath Get the full path to the migration.
func (r *Creator) GetPath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, support.Config.Paths.Migration, name+".go")
}

// GetFileName Get the full path to the migration.
func (r *Creator) GetFileName(name string) string {
	return fmt.Sprintf("%s_%s", carbon.Now().ToShortDateTimeString(), name)
}
