package modify

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/spf13/cast"
	"golang.org/x/tools/imports"

	"github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
)

func Env(file, key, value string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		document := parseEnvFile(string(content))
		updated := false
		for i := len(document.keys) - 1; i >= 0; i-- {
			entry := document.keys[i]
			if entry.key != key {
				continue
			}

			document.lines[entry.line] = entry.prefix + key + "=" + value
			if entry.end != entry.line {
				document.lines = append(document.lines[:entry.line+1], document.lines[entry.end+1:]...)
			}
			updated = true
		}
		if !updated {
			document.lines = append(document.lines, key+"="+value)
		}

		return os.WriteFile(file, []byte(document.String()), 0o644)
	})
}

func EnvWhenMissing(file, key, value string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		document := parseEnvFile(string(content))
		for _, entry := range document.keys {
			if entry.key == key {
				return nil
			}
		}

		document.lines = append(document.lines, key+"="+value)

		return os.WriteFile(file, []byte(document.String()), 0o644)
	})
}

// envKey is an assignment found in an environment file: the lines it spans,
// its key, and the "export " prefix to keep when the line is rewritten. A
// quoted value may span several lines.
type envKey struct {
	line   int
	end    int
	key    string
	prefix string
}

// envLine parses one line of an environment file with the grammar the dotenv
// parser accepts: an optional "export" prefix, a key, then "=" or ": " and the
// value. Commented and empty lines are not assignments.
func envLine(line string) (key, prefix, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", "", false
	}

	if rest, found := strings.CutPrefix(line, "export"); found && rest != "" && isEnvSpace(rest[0]) {
		line = strings.TrimLeft(rest, " \t")
		prefix = "export "
	}

	index := strings.IndexAny(line, "=:")
	if index == -1 {
		return "", "", "", false
	}
	if line[index] == ':' && (index+1 == len(line) || !isEnvSpace(line[index+1])) {
		// The dotenv parser requires whitespace after a colon.
		return "", "", "", false
	}

	return strings.TrimSpace(line[:index]), prefix, line[index+1:], true
}

func isEnvSpace(char byte) bool {
	return char == ' ' || char == '\t'
}

// envFile is an environment file split into lines, with the byte order mark,
// the line ending and the final newline it uses, so a write keeps the shape of
// the file.
type envFile struct {
	bom      bool
	lines    []string
	keys     []envKey
	eol      string
	trailing bool
}

// parseEnvFile splits content into lines and records the assignments outside
// quoted values, so a rewrite touches only real assignments.
func parseEnvFile(content string) envFile {
	document := envFile{eol: "\n"}
	if strings.HasPrefix(content, "\ufeff") {
		document.bom = true
		content = strings.TrimPrefix(content, "\ufeff")
	}
	if strings.Contains(content, "\r\n") {
		document.eol = "\r\n"
	}

	if content == "" {
		document.trailing = true

		return document
	}

	if strings.HasSuffix(content, "\n") {
		document.trailing = true
		content = strings.TrimSuffix(content, "\n")
	}

	document.lines = strings.Split(content, "\n")
	for i, line := range document.lines {
		document.lines[i] = strings.TrimSuffix(line, "\r")
	}
	document.keys = envKeys(document.lines)

	return document
}

// envKeys returns the assignments of an environment file. A quoted value may
// span several lines, and a line that looks like an assignment inside it is not
// one.
func envKeys(lines []string) []envKey {
	var keys []envKey
	var quote byte

	for i, line := range lines {
		if quote != 0 {
			if closesEnvQuote(line, quote) {
				quote = 0
				keys[len(keys)-1].end = i
			}

			continue
		}

		key, prefix, value, ok := envLine(line)
		if !ok {
			continue
		}

		keys = append(keys, envKey{line: i, end: i, key: key, prefix: prefix})

		if next, open := opensEnvQuote(value); open {
			quote = next
		}
	}

	return keys
}

// opensEnvQuote reports whether value starts a quoted value that does not close
// on the same line.
func opensEnvQuote(value string) (byte, bool) {
	value = strings.TrimSpace(value)
	if value == "" || (value[0] != '"' && value[0] != '\'') {
		return 0, false
	}

	return value[0], !closesEnvQuote(value[1:], value[0])
}

// closesEnvQuote reports whether line holds the closing quote, ignoring an
// escaped one.
func closesEnvQuote(line string, quote byte) bool {
	for i := 0; i < len(line); i++ {
		if line[i] == quote && (i == 0 || line[i-1] != '\\') {
			return true
		}
	}

	return false
}

// String rebuilds the file with the byte order mark, line ending and final
// newline it had.
func (f envFile) String() string {
	content := strings.Join(f.lines, f.eol)
	if f.trailing {
		content += f.eol
	}
	if f.bom {
		content = "\ufeff" + content
	}

	return content
}

func RegisterMigration(pkg, migration string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return AddMigration(pkg, migration)
	})
}

func RegisterProvider(pkg, provider string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return AddProvider(pkg, provider)
	})
}

func RegisterRoute(pkg, route string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return AddRoute(pkg, route)
	})
}

func Call(fn func(options []modify.Option) error) modify.Apply {
	return &callModifier{
		fn: fn,
	}
}

func File(path string) modify.File {
	return &file{path: path}
}

func GoFile(file string) modify.GoFile {
	return &goFile{file: file}
}

func UnregisterMigration(pkg, migration string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return RemoveMigration(pkg, migration)
	})
}

func UnregisterProvider(pkg, provider string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return RemoveProvider(pkg, provider)
	})
}

func UnregisterRoute(pkg, route string) modify.Apply {
	return Call(func(_ []modify.Option) error {
		return RemoveRoute(pkg, route)
	})
}

func When(fn func(options map[string]any) bool, applies ...modify.Apply) modify.Apply {
	return &whenModifier{
		fn:      fn,
		applies: applies,
	}
}

func WhenDriver(driver string, applies ...modify.Apply) modify.Apply {
	return When(func(options map[string]any) bool {
		return driver == options["driver"]
	}, applies...)
}

func WhenFileExists(file string, applies ...modify.Apply) modify.Apply {
	return Call(func(options []modify.Option) error {
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				return nil
			}

			return err
		}

		for _, apply := range applies {
			if err := apply.Apply(options...); err != nil {
				return err
			}
		}

		return nil
	})
}

func WhenFacade(facade string, applies ...modify.Apply) modify.Apply {
	return When(func(options map[string]any) bool {
		return facade == options["facade"]
	}, applies...)
}

func WhenFileContains(file, content string, applies ...modify.Apply) modify.Apply {
	return When(func(_ map[string]any) bool {
		return supportfile.Contains(file, content)
	}, applies...)
}

func WhenFileNotContains(file, content string, applies ...modify.Apply) modify.Apply {
	return When(func(_ map[string]any) bool {
		return !supportfile.Contains(file, content)
	}, applies...)
}

func WhenFileNotExists(file string, applies ...modify.Apply) modify.Apply {
	return When(func(_ map[string]any) bool {
		return !supportfile.Exists(file)
	}, applies...)
}

func WhenNoFacades(facades []string, applies ...modify.Apply) modify.Apply {
	return When(func(options map[string]any) bool {
		var exist bool
		for _, facade := range facades {
			if facade == options["facade"] {
				continue
			}

			if supportfile.Exists(facadeToFilepath(facade)) {
				exist = true
				break
			}
		}

		return !exist
	}, applies...)
}

func generateOptions(options []modify.Option) map[string]any {
	result := make(map[string]any)
	for _, option := range options {
		option(result)
	}
	return result
}

type callModifier struct {
	fn func(options []modify.Option) error
}

func (r *callModifier) Apply(options ...modify.Option) error {
	return r.fn(options)
}

type file struct {
	path string
}

func (r *file) Append(content string) modify.Apply {
	return &appendFile{
		content: content,
		path:    r.path,
	}
}

func (r *file) Overwrite(content string) modify.Apply {
	return &overwriteFile{
		content: content,
		path:    r.path,
	}
}

func (r *file) Remove() modify.Apply {
	return &removeFile{
		path: r.path,
	}
}

type appendFile struct {
	content string
	path    string
}

func (r *appendFile) Apply(options ...modify.Option) error {
	return supportfile.PutContent(r.path, r.content, supportfile.WithAppend())
}

type overwriteFile struct {
	content string
	path    string
}

func (r *overwriteFile) Apply(options ...modify.Option) error {
	generatedOptions := generateOptions(options)

	if supportfile.Exists(r.path) && !cast.ToBool(generatedOptions["force"]) {
		color.Warningln(errors.ConsoleFileAlreadyExists.Args(r.path))
		return nil
	}

	return supportfile.PutContent(r.path, r.content)
}

type removeFile struct {
	path string
}

func (r *removeFile) Apply(...modify.Option) error {
	return supportfile.Remove(r.path)
}

type goFile struct {
	file      string
	modifiers []modify.GoNode
}

func (r goFile) Apply(...modify.Option) error {
	source, err := supportfile.GetContent(r.file)
	if err != nil {
		return err
	}

	df, err := decorator.Parse(source)
	if err != nil {
		return err
	}

	for i := range r.modifiers {
		if err = r.modifiers[i].Apply(df); err != nil {
			return errors.PackageModifyGoFileFail.Args(r.file, err)
		}
	}

	var buf bytes.Buffer
	err = decorator.Fprint(&buf, df)
	if err != nil {
		return err
	}

	return supportfile.PutContent(r.file, buf.String())
}

func (r goFile) Find(matchers []match.GoNode) modify.GoNode {
	modifier := &goNode{
		matchers: matchers,
		goFile:   &r,
	}
	r.modifiers = append(r.modifiers, modifier)

	return modifier
}

func (r goFile) FindOrCreate(matchers []match.GoNode, fn func(node dst.Node) error) modify.GoNode {
	modifier := &goNode{
		createFunc: fn,
		matchers:   matchers,
		goFile:     &r,
	}
	r.modifiers = append(r.modifiers, modifier)

	return modifier
}

func (r goFile) Format() modify.Apply {
	return Call(func(options []modify.Option) error {
		if err := r.Apply(options...); err != nil {
			return err
		}

		content, err := supportfile.GetContent(r.file)
		if err != nil {
			return err
		}

		// Clear unused imports
		opts := &imports.Options{
			TabWidth:  8,
			TabIndent: true,
			Comments:  true,
			Fragment:  false,
		}
		formatted, err := imports.Process(filepath.Base(r.file), []byte(content), opts)
		if err != nil {
			return err
		}

		// Format source code
		formatted, err = format.Source(formatted)
		if err != nil {
			return err
		}

		return supportfile.PutContent(r.file, string(formatted))
	})
}

type goNode struct {
	actions    []modify.Action
	createFunc func(node dst.Node) error
	goFile     *goFile
	matchers   []match.GoNode
}

func (r *goNode) Apply(node dst.Node) (err error) {
	var (
		current      int
		matched      bool
		matchedNodes = make(map[dst.Node]bool)
	)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// match the node and apply the action
	dstutil.Apply(node, func(cursor *dstutil.Cursor) bool {
		// if already modified, skip the rest of the nodes
		if matched {
			return false
		}

		if r.matchers[current].MatchCursor(cursor) {
			matchedNodes[cursor.Node()] = true
			current++
			if current == len(r.matchers) {
				// apply the actions after all matchers are matched
				for _, action := range r.actions {
					action(cursor)
				}
				matched = true

				return false
			}
		}

		return true
	}, func(cursor *dstutil.Cursor) bool {
		if nd := cursor.Node(); nd != nil && matchedNodes[nd] {
			return false
		}

		return true
	})

	if !matched {
		if r.createFunc != nil {
			if err := r.createFunc(node); err != nil {
				return err
			}

			r.createFunc = nil // prevent infinite recursion

			return r.Apply(node) // try to apply again after creation
		}

		count := len(r.matchers)
		return errors.PackageMatchGoNodeFail.Args(count-current, count)
	}

	return nil
}

func (r *goNode) Modify(actions ...modify.Action) modify.GoFile {
	r.actions = actions

	return r.goFile
}

type whenModifier struct {
	fn      func(options map[string]any) bool
	applies []modify.Apply
}

func (r whenModifier) Apply(options ...modify.Option) error {
	if !r.fn(generateOptions(options)) {
		return nil
	}

	for _, apply := range r.applies {
		if err := apply.Apply(options...); err != nil {
			return err
		}
	}

	return nil
}

func facadeToFilepath(facade string) string {
	return path.Facade(str.Of(facade).Snake().String() + ".go")
}
