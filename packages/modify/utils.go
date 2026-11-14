package modify

import (
	"cmp"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support/env"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

// consolePackageImportPath is the contracts/console import the command slices are
// typed with, shared by AddCommand and the WithCommandHooks helpers.
const consolePackageImportPath = "github.com/goravel/framework/contracts/console"

// AddCommand adds command to the foundation.Setup() chain in the Boot function.
func AddCommand(pkg, command string) error {
	config := withSliceConfig{
		fileName:        "commands.go",
		withMethodName:  "WithCommands",
		helperFuncName:  "Commands",
		typePackage:     "console",
		typeName:        "Command",
		typeImportPath:  consolePackageImportPath,
		fileExistsError: errors.PackageCommandsFileExists,
		stubTemplate:    commands,
		matcherFunc:     match.Commands,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, command)
}

// AddCommandHook registers a hook in bootstrap/app.go under a command signature
// or '*' pattern, e.g. "migrate" or "make:*". It creates the WithCommandHooks
// call when missing and resolves a named argument to its declaration in the
// bootstrap package. A file without a foundation.Setup() chain returns false
// without an error.
func AddCommandHook(pkg, key, hook string) (bool, error) {
	appFilePath := path.Bootstrap("app.go")

	if !supportfile.Contains(appFilePath, "foundation.Setup().") {
		return false, nil
	}

	consoleName := "console"
	registered := false
	namedHook := ""

	if err := GoFile(appFilePath).
		FindOrCreate(match.Imports(), createImport).
		Modify(captureImportName(consolePackageImportPath, &consoleName)).
		Find(match.FoundationSetup()).
		Modify(foundationSetupCommandHook(key, hook, &consoleName, &registered, &namedHook)).
		FindOrCreate(match.Imports(), createImport).
		Modify(addCommandHookImports(pkg, &registered)).
		Apply(); err != nil {
		return false, err
	}

	if namedHook != "" {
		registered, err := appendToNamedCommandHook(filepath.Dir(appFilePath), namedHook, pkg, key, hook)
		if err != nil {
			return false, err
		}
		if !registered {
			return false, errors.PackageCommandHookNotRegistered.Args(hook)
		}

		return true, nil
	}

	if !registered {
		return false, errors.PackageCommandHookNotRegistered.Args(hook)
	}

	return true, nil
}

// AddFilter adds filter to the foundation.Setup() chain in the Boot function.
func AddFilter(pkg, filter string) error {
	config := withSliceConfig{
		fileName:        "filters.go",
		withMethodName:  "WithFilters",
		helperFuncName:  "Filters",
		typePackage:     "validation",
		typeName:        "Filter",
		typeImportPath:  "github.com/goravel/framework/contracts/validation",
		fileExistsError: errors.PackageFiltersFileExists,
		stubTemplate:    filters,
		matcherFunc:     match.Filters,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, filter)
}

// AddJob adds job to the foundation.Setup() chain in the Boot function.
func AddJob(pkg, job string) error {
	config := withSliceConfig{
		fileName:        "jobs.go",
		withMethodName:  "WithJobs",
		helperFuncName:  "Jobs",
		typePackage:     "queue",
		typeName:        "Job",
		typeImportPath:  "github.com/goravel/framework/contracts/queue",
		fileExistsError: errors.PackageJobsFileExists,
		stubTemplate:    jobs,
		matcherFunc:     match.Jobs,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, job)
}

// AddMiddleware adds middleware to the foundation.Setup() chain in the Boot function.
func AddMiddleware(pkg, middleware string) error {
	appFilePath := path.Bootstrap("app.go")

	if err := addMiddlewareImports(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupMiddleware(middleware)).Apply()
}

// AddMigration adds migration to the foundation.Setup() chain in the Boot function.
func AddMigration(pkg, migration string) error {
	config := withSliceConfig{
		fileName:        "migrations.go",
		withMethodName:  "WithMigrations",
		helperFuncName:  "Migrations",
		typePackage:     "schema",
		typeName:        "Migration",
		typeImportPath:  "github.com/goravel/framework/contracts/database/schema",
		fileExistsError: errors.PackageMigrationsFileExists,
		stubTemplate:    migrations,
		matcherFunc:     match.Migrations,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, migration)
}

// AddModule adds module to bootstrap/modules.go, wiring WithModules(Modules) when missing.
// An existing modules.go is the list, even before WithModules is wired by hand.
func AddModule(pkg, module string) error {
	handler := newWithSliceHandler(moduleConfig())
	if handler.fileExists {
		return handler.addItemToFile(pkg, module)
	}

	return handler.AddItem(pkg, module)
}

// moduleMethods are the module capabilities AddToModule can edit: method -> slice type.
var moduleMethods = map[string]struct{ typeName, typeImportPath string }{
	"Commands":         {"console.Command", "github.com/goravel/framework/contracts/console"},
	"Jobs":             {"queue.Job", "github.com/goravel/framework/contracts/queue"},
	"Migrations":       {"schema.Migration", "github.com/goravel/framework/contracts/database/schema"},
	"Seeders":          {"seeder.Seeder", "github.com/goravel/framework/contracts/database/seeder"},
	"ServiceProviders": {"foundation.ServiceProvider", "github.com/goravel/framework/contracts/foundation"},
}

// AddToModule adds item to the slice returned by method of the module struct in dir
// (relative to the project root), appending the method when the struct lacks it.
func AddToModule(dir, method, pkg, item string) error {
	spec, ok := moduleMethods[method]
	if !ok {
		return errors.PackageModuleMethodUnsupported.Args(method)
	}

	target, receiver, exists, err := findModuleMethod(dir, method)
	if err != nil {
		return err
	}

	if !exists {
		stub := "\nfunc " + receiver + " " + method + "() []" + spec.typeName + " {\n\treturn []" + spec.typeName + "{}\n}\n"
		if err := supportfile.PutContent(target, stub, supportfile.WithAppend()); err != nil {
			return err
		}
		if err := addImportsToFile(target, spec.typeImportPath); err != nil {
			return err
		}
	}

	if err := addImportsToFile(target, pkg); err != nil {
		return err
	}

	// The application's Go module is what pkg has before the module directory.
	main := env.MainPath()
	if _, importPath, found := strings.Cut(pkg, " "); found {
		pkg = importPath
	}
	if before, _, found := strings.Cut(pkg, "/"+strings.Trim(filepath.ToSlash(dir), "/")+"/"); found {
		main = before
	}

	return GoFile(target).
		Find(match.ModuleMethod(method)).Modify(Register(item)).
		Find(match.Imports()).Modify(groupImports(main)).
		Apply()
}

// AddProvider adds service provider to the foundation.Setup() chain in the Boot function.
func AddProvider(pkg, provider string) error {
	config := withSliceConfig{
		fileName:        "providers.go",
		withMethodName:  "WithProviders",
		helperFuncName:  "Providers",
		typePackage:     "foundation",
		typeName:        "ServiceProvider",
		typeImportPath:  "github.com/goravel/framework/contracts/foundation",
		fileExistsError: errors.PackageProvidersFileExists,
		stubTemplate:    providers,
		matcherFunc:     match.Providers,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, provider)
}

// AddRoute adds route to the foundation.Setup() chain in the Boot function.
// Add WithRouting(func()) to foundation.Setup() if not exists.
// Add pkg to this file imports, and add route to the function body in WithRouting.
// the pkg is the package path of the route file, e.g., "goravel/routes"
// the route will be like "routes.Web()"
func AddRoute(pkg, route string) error {
	appFilePath := path.Bootstrap("app.go")

	if err := addImportsToFile(appFilePath, pkg); err != nil {
		return err
	}

	return GoFile(appFilePath).Find(match.FoundationSetup()).Modify(foundationSetupRouting(route)).Apply()
}

// AddRule adds rule to the foundation.Setup() chain in the Boot function.
func AddRule(pkg, rule string) error {
	config := withSliceConfig{
		fileName:        "rules.go",
		withMethodName:  "WithRules",
		helperFuncName:  "Rules",
		typePackage:     "validation",
		typeName:        "Rule",
		typeImportPath:  "github.com/goravel/framework/contracts/validation",
		fileExistsError: errors.PackageRulesFileExists,
		stubTemplate:    rules,
		matcherFunc:     match.Rules,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, rule)
}

// AddSeeder adds seeder to the foundation.Setup() chain in the Boot function.
func AddSeeder(pkg, seeder string) error {
	config := withSliceConfig{
		fileName:        "seeders.go",
		withMethodName:  "WithSeeders",
		helperFuncName:  "Seeders",
		typePackage:     "seeder",
		typeName:        "Seeder",
		typeImportPath:  "github.com/goravel/framework/contracts/database/seeder",
		fileExistsError: errors.PackageSeedersFileExists,
		stubTemplate:    seeders,
		matcherFunc:     match.Seeders,
	}

	handler := newWithSliceHandler(config)
	return handler.AddItem(pkg, seeder)
}

// ExprExists checks if an expression exists in a slice of expressions.
// It uses structural equality comparison via ExprIndex.
//
// Parameters:
//   - x: Slice of expressions to search in
//   - y: Expression to search for
//
// Returns true if the expression exists in the slice, false otherwise.
//
// Example:
//
//	exprs := []dst.Expr{
//		&dst.Ident{Name: "foo"},
//		&dst.Ident{Name: "bar"},
//	}
//	target := &dst.Ident{Name: "foo"}
//	if ExprExists(exprs, target) {
//		fmt.Println("Expression found")
//	}
func ExprExists(x []dst.Expr, y dst.Expr) bool {
	return ExprIndex(x, y) >= 0
}

// ExprIndex returns the index of the first occurrence of an expression in a slice.
// It uses structural equality comparison to match expressions.
//
// Parameters:
//   - x: Slice of expressions to search in
//   - y: Expression to search for
//
// Returns the index of the first occurrence, or -1 if not found.
//
// Example:
//
//	exprs := []dst.Expr{
//		&dst.Ident{Name: "foo"},
//		&dst.Ident{Name: "bar"},
//		&dst.Ident{Name: "baz"},
//	}
//	target := &dst.Ident{Name: "bar"}
//	index := ExprIndex(exprs, target) // returns 1
func ExprIndex(x []dst.Expr, y dst.Expr) int {
	return slices.IndexFunc(x, func(expr dst.Expr) bool {
		return match.EqualNode(y).MatchNode(expr)
	})
}

// InitModules creates bootstrap/modules.go and wires WithModules(Modules), with no module listed.
func InitModules() error {
	return newWithSliceHandler(moduleConfig()).EnsureWithMethod()
}

// ModulesStub is the content of an empty bootstrap/modules.go.
func ModulesStub() string {
	return modules()
}

func moduleConfig() withSliceConfig {
	return withSliceConfig{
		fileName:        "modules.go",
		withMethodName:  "WithModules",
		helperFuncName:  "Modules",
		typePackage:     "modular",
		typeName:        "Module",
		typeImportPath:  "github.com/goravel/framework/contracts/modular",
		fileExistsError: errors.PackageModulesFileExists,
		stubTemplate:    modules,
		matcherFunc:     match.Modules,
	}
}

// IsUsingImport checks if an imported package is actually used in the file.
// It inspects the AST for selector expressions that reference the package.
//
// Parameters:
//   - df: The parsed Go file to inspect
//   - path: Import path of the package (e.g., "github.com/goravel/framework/contracts/console")
//   - name: Optional package name. If not provided, uses the last segment of the path
//
// Returns true if the package is used anywhere in the file, false otherwise.
//
// Example:
//
//	file, _ := decorator.Parse(src)
//	// Check if "console" package from "github.com/goravel/framework/contracts/console" is used
//	if IsUsingImport(file, "github.com/goravel/framework/contracts/console") {
//		fmt.Println("console package is being used")
//	}
//	// Or specify a custom name
//	if IsUsingImport(file, "github.com/goravel/framework/contracts/console", "customName") {
//		fmt.Println("Package with alias 'customName' is being used")
//	}
func IsUsingImport(df *dst.File, path string, name ...string) bool {
	if len(name) == 0 {
		split := strings.Split(path, "/")
		name = append(name, split[len(split)-1])
	}

	var used bool
	dst.Inspect(df, func(n dst.Node) bool {
		sel, ok := n.(*dst.SelectorExpr)
		if ok && isTopName(sel.X, name[0]) {
			used = true

			return false
		}
		return true
	})

	return used
}

// KeyExists checks if a key exists in a slice of key-value expressions.
// It uses structural equality comparison via KeyIndex.
//
// Parameters:
//   - kvs: Slice of expressions (expected to contain KeyValueExpr)
//   - key: Key expression to search for
//
// Returns true if the key exists in any KeyValueExpr, false otherwise.
//
// Example:
//
//	kvExprs := []dst.Expr{
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "name"},
//			Value: &dst.BasicLit{Value: `"John"`},
//		},
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "age"},
//			Value: &dst.BasicLit{Value: "30"},
//		},
//	}
//	targetKey := &dst.Ident{Name: "name"}
//	if KeyExists(kvExprs, targetKey) {
//		fmt.Println("Key found")
//	}
func KeyExists(kvs []dst.Expr, key dst.Expr) bool {
	return KeyIndex(kvs, key) >= 0
}

// KeyIndex returns the index of a key in a slice of key-value expressions.
// It searches for KeyValueExpr nodes and compares their keys using structural equality.
//
// Parameters:
//   - kvs: Slice of expressions (expected to contain KeyValueExpr)
//   - key: Key expression to search for
//
// Returns the index of the first KeyValueExpr with matching key, or -1 if not found.
//
// Example:
//
//	kvExprs := []dst.Expr{
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "name"},
//			Value: &dst.BasicLit{Value: `"John"`},
//		},
//		&dst.KeyValueExpr{
//			Key:   &dst.Ident{Name: "age"},
//			Value: &dst.BasicLit{Value: "30"},
//		},
//	}
//	targetKey := &dst.Ident{Name: "age"}
//	index := KeyIndex(kvExprs, targetKey) // returns 1
func KeyIndex(kvs []dst.Expr, key dst.Expr) int {
	return slices.IndexFunc(kvs, func(expr dst.Expr) bool {
		if kv, ok := expr.(*dst.KeyValueExpr); ok {
			return match.EqualNode(key).MatchNode(kv.Key)
		}
		return false
	})
}

// MustParseExpr parses a Go expression from a string and returns its AST node.
// It wraps the expression in a minimal valid Go program to parse it, then extracts
// and returns the expression node with proper decorations and newlines.
//
// Parameters:
//   - x: String representation of a Go expression
//
// Returns the parsed expression as a dst.Node, with decorations preserved.
// Panics if the expression cannot be parsed.
//
// Example:
//
//	// Parse a simple expression
//	node := MustParseExpr("&commands.ExampleCommand{}")
//	// Returns a UnaryExpr node representing the address-of operation
//
//	// Parse a composite literal
//	node := MustParseExpr(`map[string]interface{}{"key": "value"}`)
//	// Returns a CompositeLit node
//
//	// Parse a function call
//	node := MustParseExpr("fmt.Println(\"hello\")")
//	// Returns a CallExpr node
func MustParseExpr(x string) (node dst.Node) {
	src := "package p\nvar _ = " + x
	file, err := decorator.Parse(src)
	if err != nil {
		panic(err)
	}

	spec := file.Decls[0].(*dst.GenDecl).Specs[0].(*dst.ValueSpec)
	expr := spec.Values[0]

	// handle outer comments for expr
	expr.Decorations().Start = file.Decls[0].(*dst.GenDecl).Decorations().Start
	expr.Decorations().End = file.Decls[0].(*dst.GenDecl).Decorations().End

	return WrapNewline(expr)
}

func RemoveMigration(pkg, migration string) error {
	config := withSliceConfig{
		fileName:       "migrations.go",
		withMethodName: "WithMigrations",
		helperFuncName: "Migrations",
		typePackage:    "schema",
		typeName:       "Migration",
		typeImportPath: "github.com/goravel/framework/contracts/database/schema",
		matcherFunc:    match.Migrations,
	}

	handler := newWithSliceHandler(config)
	return handler.RemoveItem(pkg, migration)
}

// RemoveModule removes module from the Modules() list.
func RemoveModule(pkg, module string) error {
	return newWithSliceHandler(moduleConfig()).RemoveItem(pkg, module)
}

// RemoveProvider removes a service provider from the foundation.Setup() chain in the Boot function.
func RemoveProvider(pkg, provider string) error {
	config := withSliceConfig{
		fileName:       "providers.go",
		withMethodName: "WithProviders",
		helperFuncName: "Providers",
		typePackage:    "foundation",
		typeName:       "ServiceProvider",
		typeImportPath: "github.com/goravel/framework/contracts/foundation",
		matcherFunc:    match.Providers,
	}

	handler := newWithSliceHandler(config)
	return handler.RemoveItem(pkg, provider)
}

// RemoveRoute removes a route from the foundation.Setup() chain in the Boot function.
func RemoveRoute(pkg, route string) error {
	appFilePath := path.Bootstrap("app.go")

	if err := GoFile(appFilePath).Find(match.FoundationSetup()).Modify(removeRouteFromSetup(route)).Apply(); err != nil {
		return err
	}

	return removeImportsFromFile(appFilePath, pkg)
}

// WrapNewline adds newline decorations to specific AST nodes for better formatting.
// It traverses the AST and adds Before/After newlines to KeyValueExpr, UnaryExpr,
// and FuncType result nodes to improve code readability.
//
// Parameters:
//   - node: Any dst.Node to process
//
// Returns the same node with newline decorations applied.
//
// Example:
//
//	// Parse and wrap an expression
//	expr := MustParseExpr("&commands.ExampleCommand{}")
//	// The UnaryExpr will have newlines before and after
//
//	// For a composite literal with key-value pairs:
//	node := MustParseExpr(`map[string]int{"a": 1, "b": 2}`)
//	wrapped := WrapNewline(node)
//	// Each KeyValueExpr will have newlines for better formatting:
//	// map[string]int{
//	//     "a": 1,
//	//     "b": 2,
//	// }
func WrapNewline[T dst.Node](node T) T {
	dst.Inspect(node, func(n dst.Node) bool {
		switch v := n.(type) {
		case *dst.KeyValueExpr, *dst.UnaryExpr:
			v.Decorations().After = dst.NewLine
			v.Decorations().Before = dst.NewLine
		case *dst.FuncType:
			v.Results.Decorations().After = dst.NewLine
			v.Results.Decorations().Before = dst.NewLine
		}

		return true
	})

	return node
}

// groupImports sorts the matched import block into standard library, third-party
// and application (inside the Go module main) groups, separated by an empty line.
func groupImports(main string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		groupImportSpecs(cursor.Node().(*dst.GenDecl), main)
	}
}

func groupImportSpecs(node *dst.GenDecl, main string) {
	importPath := func(spec dst.Spec) string {
		value, _ := strconv.Unquote(spec.(*dst.ImportSpec).Path.Value)

		return value
	}
	group := func(spec dst.Spec) int {
		value := importPath(spec)
		switch {
		case value == main || strings.HasPrefix(value, main+"/"):
			return 2
		case isThirdParty(value):
			return 1
		default:
			return 0
		}
	}

	slices.SortStableFunc(node.Specs, func(a, b dst.Spec) int {
		return cmp.Or(cmp.Compare(group(a), group(b)), cmp.Compare(importPath(a), importPath(b)))
	})
	for i, spec := range node.Specs {
		decs := spec.Decorations()
		decs.Before, decs.After = dst.NewLine, dst.NewLine
		if i > 0 && group(spec) != group(node.Specs[i-1]) {
			decs.Before = dst.EmptyLine
		}
	}
}

// findModuleMethod returns the file declaring method on the module struct of dir. When
// no file does, it returns the file declaring Name() and the receiver to declare it with.
func findModuleMethod(dir, method string) (target, receiver string, exists bool, err error) {
	root := path.Base(filepath.FromSlash(dir))
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		return "", "", false, err
	}

	// module.go first: it wins when several types have a Name() method.
	slices.SortStableFunc(files, func(a, b string) int {
		return cmp.Compare(moduleFileRank(a), moduleFileRank(b))
	})

	type moduleFunc struct{ file, name, typeName, receiver string }
	var methods []moduleFunc
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
		if err != nil {
			return "", "", false, err
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
				continue
			}

			typeName, text := receiverOf(fn.Recv.List[0])
			if typeName != "" {
				methods = append(methods, moduleFunc{file: file, name: fn.Name.Name, typeName: typeName, receiver: text})
			}
		}
	}

	nameIndex := slices.IndexFunc(methods, func(m moduleFunc) bool { return m.name == "Name" })
	if nameIndex < 0 {
		return "", "", false, errors.PackageModuleStructNotFound.Args(dir)
	}

	module := methods[nameIndex]
	for _, m := range methods {
		if m.name == method && m.typeName == module.typeName {
			return m.file, m.receiver, true, nil
		}
	}

	return module.file, module.receiver, false, nil
}

func moduleFileRank(file string) int {
	if filepath.Base(file) == "module.go" {
		return 0
	}

	return 1
}

// receiverOf returns the type name of a method receiver and its source, e.g. "(m *Module)".
func receiverOf(field *ast.Field) (typeName, text string) {
	expr, pointer := field.Type, ""
	if star, ok := expr.(*ast.StarExpr); ok {
		expr, pointer = star.X, "*"
	}

	ident, ok := expr.(*ast.Ident)
	if !ok {
		return "", ""
	}

	name := ""
	if len(field.Names) == 1 {
		name = field.Names[0].Name + " "
	}

	return ident.Name, "(" + name + pointer + ident.Name + ")"
}

// addImportsToFile adds the specified import to the file, creating the import block if needed.
func addImportsToFile(filePath, pkg string) error {
	importMatchers := match.Imports()

	if alias, importPath, found := strings.Cut(pkg, " "); found {
		return GoFile(filePath).FindOrCreate(importMatchers, createImport).Modify(AddImport(importPath, alias)).Apply()
	}

	return GoFile(filePath).FindOrCreate(importMatchers, createImport).Modify(AddImport(pkg)).Apply()
}

// isThirdParty determines if an import path refers to a third-party package.
// Third-party packages typically contain a domain (e.g., ".com", ".org") in their path.
// This heuristic is taken from golang.org/x/tools/imports package.
//
// Parameters:
//   - importPath: The import path to check
//
// Returns true if the import path appears to be a third-party package, false for standard library.
//
// Example:
//
//	isThirdParty("fmt") // false - standard library
//	isThirdParty("encoding/json") // false - standard library
//	isThirdParty("github.com/goravel/framework") // true - third party
//	isThirdParty("example.com/package") // true - third party
func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isTopName checks if an expression is a top-level unresolved identifier with the given name.
// An identifier is considered "top-level" and "unresolved" when it has no associated object,
// meaning it refers to a package name or other non-local identifier.
//
// Parameters:
//   - n: Expression to check
//   - name: Expected identifier name
//
// Returns true if n is an Ident with the given name and no associated object, false otherwise.
//
// Example:
//
//	// In the expression "fmt.Println", "fmt" is a top-level identifier
//	selectorExpr := &dst.SelectorExpr{
//		X:   &dst.Ident{Name: "fmt", Obj: nil},
//		Sel: &dst.Ident{Name: "Println"},
//	}
//	isTopName(selectorExpr.X, "fmt") // true
//
//	// A local variable has an associated Obj, so it's not a top-level name
//	localVar := &dst.Ident{Name: "x", Obj: &dst.Object{}}
//	isTopName(localVar, "x") // false
func isTopName(n dst.Expr, name string) bool {
	id, ok := n.(*dst.Ident)
	return ok && id.Name == name && id.Obj == nil
}

// removeImportsFromFile removes the specified import from the file.
func removeImportsFromFile(filePath, pkg string) error {
	importMatchers := match.Imports()

	if alias, importPath, found := strings.Cut(pkg, " "); found {
		return GoFile(filePath).Find(importMatchers).Modify(RemoveImport(importPath, alias)).Apply()
	}

	return GoFile(filePath).Find(importMatchers).Modify(RemoveImport(pkg)).Apply()
}
