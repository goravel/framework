package console

import (
	"fmt"
	"go/format"
	"path/filepath"
	"regexp"
	"strconv"
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

type modelInfo struct {
	Fields          []string
	Embeds          []string
	Imports         map[string]struct{}
	NeedsTableName  bool
	TableName       string
	TableNameMethod string
}

type fieldInfo struct {
	Name    string
	Type    string
	Tags    string
	Imports []string
}

type ModelMakeCommand struct {
	artisan console.Artisan
	schema  schema.Schema
	grammar driver.Grammar
}

func NewModelMakeCommand(artisan console.Artisan, schema schema.Schema, grammar driver.Grammar) *ModelMakeCommand {
	return &ModelMakeCommand{
		artisan: artisan,
		schema:  schema,
		grammar: grammar,
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
	modelInfo := modelInfo{}
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

		modelInfo, err = r.generateModelInfo(columns, structName, table)
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
		ctx.Info(fmt.Sprintf("Generated %d fields and %d embeds from table '%s'", len(modelInfo.Fields), len(modelInfo.Embeds), table))
	}

	stubContent, err := r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), modelInfo)
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

func (r *ModelMakeCommand) generateModelInfo(columns []driver.Column, structName, tableName string) (modelInfo, error) {
	info := modelInfo{
		Imports:   make(map[string]struct{}),
		Fields:    []string{},
		Embeds:    []string{},
		TableName: tableName,
	}

	var hasID, hasCreatedAt, hasUpdatedAt, hasDeletedAt bool
	for _, col := range columns {
		switch col.Name {
		case "id":
			hasID = true
		case "created_at":
			hasCreatedAt = true
		case "updated_at":
			hasUpdatedAt = true
		case "deleted_at":
			hasDeletedAt = true
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

	typePatternMapping := r.grammar.TypePatternMapping()
	goTypeMapping := r.schema.GoTypeMap()

	for _, col := range columns {
		skip := false
		switch col.Name {
		case "id", "created_at", "updated_at":
			skip = embedOrmModel || embedOrmTimestamps
		case "deleted_at":
			skip = embedOrmSoftDeletes
		}
		if skip {
			continue
		}

		field := generateField(col, typePatternMapping, goTypeMapping)
		if len(field.Imports) > 0 {
			for _, importPath := range field.Imports {
				info.Imports[importPath] = struct{}{}
			}
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

func (r *ModelMakeCommand) populateStub(stub, packageName, structName string, modelInfo modelInfo) (string, error) {
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

	var buf strings.Builder
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	formatted, err := formatGoCode(buf.String())
	if err != nil {
		return "", fmt.Errorf("failed to format generated Go code: %w", err)
	}

	return formatted, nil
}

func formatGoCode(source string) (string, error) {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

func generateField(column driver.Column, typePatternMapping []driver.TypePatternMapping, goTypeMapping map[string]schema.GoTypeMapping) fieldInfo {
	schemaType := getSchemaType(column.Type, typePatternMapping)

	goType := "any"
	var imports []string

	if schemaType != "" {
		if typeInfo, ok := goTypeMapping[schemaType]; ok {
			goType = typeInfo.Type
			if column.Nullable && typeInfo.NullType != "" {
				goType = typeInfo.NullType
			}

			imports = typeInfo.Imports

			// TODO: discuss precision based typing
		}
	}

	tagParts := []string{fmt.Sprintf(`json:"%s"`, column.Name)}

	return fieldInfo{
		Name:    str.Of(column.Name).Studly().String(),
		Type:    goType,
		Tags:    "`" + strings.Join(tagParts, " ") + "`",
		Imports: imports,
	}
}

func getSchemaType(ttype string, typePatternMapping []driver.TypePatternMapping) string {
	for _, mapping := range typePatternMapping {
		if matched, err := regexp.MatchString(mapping.Pattern, ttype); err == nil && matched {
			return mapping.SchemaType
		}
	}

	return ""
}

func matchPrecisionRange(precision int, rangeStr string) bool {
	parts := strings.Split(rangeStr, ":")
	switch len(parts) {
	case 1:
		if val, err := strconv.Atoi(parts[0]); err == nil {
			return precision == val
		}
	case 2:
		start, end := parts[0], parts[1]
		if start == "" {
			if limit, err := strconv.Atoi(end); err == nil {
				return precision <= limit
			}
		} else if end == "" {
			if limit, err := strconv.Atoi(start); err == nil {
				return precision >= limit
			}
		} else {
			if s, err := strconv.Atoi(start); err == nil {
				if e, err := strconv.Atoi(end); err == nil {
					return precision >= s && precision <= e
				}
			}
		}
	}
	return false
}
