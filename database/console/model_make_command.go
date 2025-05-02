package console

import (
	"bytes"
	"fmt"
	"go/format"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/schema"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type modelDefinition struct {
	Fields          []string
	Embeds          []string
	Imports         map[string]struct{}
	NeedsTableName  bool
	TableName       string
	TableNameMethod string
}

type fieldDefinition struct {
	Name    string
	Type    string
	Tags    string
	Imports []string
}

type ModelMakeCommand struct {
	artisan console.Artisan
	schema  schema.Schema
}

func NewModelMakeCommand(artisan console.Artisan, schema schema.Schema) *ModelMakeCommand {
	return &ModelMakeCommand{
		artisan: artisan,
		schema:  schema,
	}
}

func (r *ModelMakeCommand) Signature() string {
	return "make:model"
}

func (r *ModelMakeCommand) Description() string {
	return "Create a new model class"
}

func (r *ModelMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the model even if it already exists",
			},
			&command.StringFlag{
				Name:    "table",
				Aliases: []string{"t"},
				Usage:   "Create the model from existing table schema",
			},
		},
	}
}

func (r *ModelMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "model", ctx.Argument(0), filepath.Join("app", "models"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	table := ctx.Option("table")
	model := modelDefinition{
		Imports: make(map[string]struct{}),
	}
	structName := m.GetStructName()

	if table != "" {
		if !r.schema.HasTable(table) {
			ctx.Error(fmt.Sprintf("table %s does not exist", table))
			return nil
		}

		columns, err := r.schema.GetColumns(table)
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}

		model, err = r.generateModelInfo(columns, structName, table)
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
		ctx.Info(fmt.Sprintf("Generated %d fields and %d embeds from table '%s'", len(model.Fields), len(model.Embeds), table))
	}

	stubContent, err := r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), model)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), stubContent); err != nil {
		ctx.Error(fmt.Sprintf("Error writing model file: %v", err))
		return err
	}

	ctx.Success("Model created successfully: " + m.GetFilePath())
	return nil
}

func (r *ModelMakeCommand) generateModelInfo(columns []driver.Column, structName, tableName string) (modelDefinition, error) {
	info := modelDefinition{
		Imports:   make(map[string]struct{}),
		Fields:    []string{},
		Embeds:    []string{},
		TableName: tableName,
	}

	var hasID, hasCreatedAt, hasUpdatedAt, hasDeletedAt bool
	standardColumns := make(map[string]bool)

	for _, col := range columns {
		switch col.Name {
		case "id":
			hasID = true
			standardColumns["id"] = true
		case "created_at":
			hasCreatedAt = true
			standardColumns["created_at"] = true
		case "updated_at":
			hasUpdatedAt = true
			standardColumns["updated_at"] = true
		case "deleted_at":
			hasDeletedAt = true
			standardColumns["deleted_at"] = true
		}
	}

	embedOrmModel := hasID && hasCreatedAt && hasUpdatedAt
	embedOrmTimestamps := !embedOrmModel && hasCreatedAt && hasUpdatedAt
	embedOrmSoftDeletes := hasDeletedAt

	if embedOrmModel {
		info.Embeds = append(info.Embeds, "orm.Model")
	}
	if embedOrmTimestamps {
		info.Embeds = append(info.Embeds, "orm.Timestamps")
	}
	if embedOrmSoftDeletes {
		info.Embeds = append(info.Embeds, "orm.SoftDeletes")
	}

	if len(info.Embeds) > 0 {
		info.Imports["github.com/goravel/framework/database/orm"] = struct{}{}
	}

	goTypeMapping := r.schema.GoTypes()

	for _, col := range columns {
		if (embedOrmModel || embedOrmTimestamps) &&
			(col.Name == "id" || col.Name == "created_at" || col.Name == "updated_at") {
			continue
		}

		if embedOrmSoftDeletes && col.Name == "deleted_at" {
			continue
		}

		field := generateField(col, goTypeMapping)

		for _, importPath := range field.Imports {
			info.Imports[importPath] = struct{}{}
		}

		info.Fields = append(info.Fields, r.buildField(field.Name, field.Type, field.Tags))
	}

	info.NeedsTableName = true
	info.TableNameMethod = r.buildTableNameMethod(structName, tableName)

	return info, nil
}

func (r *ModelMakeCommand) getStub() string {
	return Stubs{}.Model()
}

func (r *ModelMakeCommand) buildTableNameMethod(structName, tableName string) string {
	return fmt.Sprintf("func (r *%s) TableName() string {\n\treturn \"%s\"\n}", structName, tableName)
}

func (r *ModelMakeCommand) buildField(name, goType, tags string) string {
	return fmt.Sprintf("%-15s %-10s %s", name, goType, tags)
}

func (r *ModelMakeCommand) populateStub(stub, packageName, structName string, modelInfo modelDefinition) (string, error) {
	templateData := struct {
		PackageName     string
		StructName      string
		Embeds          []string
		Fields          []string
		TableNameMethod string
		Imports         map[string]struct{}
	}{
		PackageName:     packageName,
		StructName:      structName,
		Embeds:          modelInfo.Embeds,
		Fields:          modelInfo.Fields,
		TableNameMethod: modelInfo.TableNameMethod,
		Imports:         modelInfo.Imports,
	}

	tmpl, err := template.New("model").Parse(stub)
	if err != nil {
		return "", fmt.Errorf("failed to parse stub template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	formatted, err := formatGoCode(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to format generated Go code: %w", err)
	}

	return formatted, nil
}

func formatGoCode(source []byte) (string, error) {
	formatted, err := format.Source(source)
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

func generateField(column driver.Column, typeMapping []schema.GoTypeMapping) fieldDefinition {
	typeInfo := getSchemaType(column.Type, typeMapping)

	goType := typeInfo.Type
	if column.Nullable && typeInfo.NullType != "" {
		goType = typeInfo.NullType
	}

	tagParts := []string{fmt.Sprintf(`json:"%s"`, column.Name)}

	gormTag := fmt.Sprintf("column:%s", column.Name)
	if column.Autoincrement {
		gormTag += ";autoIncrement"
	}

	tagParts = append(tagParts, fmt.Sprintf(`gorm:"%s"`, gormTag))

	return fieldDefinition{
		Name:    str.Of(column.Name).Studly().String(),
		Type:    goType,
		Tags:    "`" + strings.Join(tagParts, " ") + "`",
		Imports: typeInfo.Imports,
	}
}

func getSchemaType(ttype string, typeMapping []schema.GoTypeMapping) schema.GoTypeMapping {
	for _, mapping := range typeMapping {
		matched, err := regexp.MatchString(mapping.Pattern, ttype)
		if err == nil && matched {
			return mapping
		}
	}

	return schema.GoTypeMapping{
		Type: "any",
	}
}
