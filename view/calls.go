package view

import (
	"cmp"
	"html/template"
	"slices"
	"text/template/parse"
)

// Call is one invocation of a template function in a compiled template set.
type Call struct {
	// Template is the name of the template the call appears in.
	Template string
	// Args are the string constants passed to the function, in order. A Dynamic
	// call still lists its constants, so {{ vite "app.js" .extra }} reports "app.js".
	Args []string
	// Dynamic reports that at least one argument is computed at render time, so
	// the call cannot be fully resolved by reading the template.
	Dynamic bool
}

// Calls returns every call of the named function in the set, ordered by template
// name then position. Nested blocks, parenthesised arguments, template invocations
// and the pipeline form {{ "x" | name }} are included. A nil set has no calls.
func Calls(tmpl *template.Template, function string) []Call {
	if tmpl == nil {
		return nil
	}

	templates := tmpl.Templates()
	slices.SortFunc(templates, func(a, b *template.Template) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	var calls []Call
	for _, t := range templates {
		if t.Tree == nil || t.Tree.Root == nil {
			continue
		}

		walkPipes(t.Tree.Root, func(pipe *parse.PipeNode) {
			for i, cmd := range pipe.Cmds {
				if !invokes(cmd, function) {
					continue
				}

				args := cmd.Args[1:]
				if i > 0 {
					// The previous command's result is passed as the final argument.
					args = append(slices.Clone(args), piped(pipe.Cmds[i-1]))
				}

				call := Call{Template: t.Name(), Args: make([]string, 0, len(args))}
				for _, arg := range args {
					str, ok := arg.(*parse.StringNode)
					if !ok {
						call.Dynamic = true
						continue
					}
					call.Args = append(call.Args, str.Text)
				}

				calls = append(calls, call)
			}
		})
	}

	return calls
}

// piped returns the value a command feeds into the next stage of its pipeline:
// its lone argument when the command is a bare literal, variable or field, or the
// command itself when the value is computed by a function call.
func piped(cmd *parse.CommandNode) parse.Node {
	if len(cmd.Args) == 1 {
		return cmd.Args[0]
	}

	return cmd
}

// invokes reports whether the command starts with the named function.
func invokes(cmd *parse.CommandNode, function string) bool {
	if len(cmd.Args) == 0 {
		return false
	}

	ident, ok := cmd.Args[0].(*parse.IdentifierNode)

	return ok && ident.Ident == function
}

// walkPipes visits every pipeline under the node.
func walkPipes(node parse.Node, visit func(*parse.PipeNode)) {
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, child := range n.Nodes {
			walkPipes(child, visit)
		}
	case *parse.ActionNode:
		walkPipes(n.Pipe, visit)
	case *parse.TemplateNode:
		if n.Pipe != nil {
			walkPipes(n.Pipe, visit)
		}
	case *parse.IfNode:
		walkBranch(&n.BranchNode, visit)
	case *parse.RangeNode:
		walkBranch(&n.BranchNode, visit)
	case *parse.WithNode:
		walkBranch(&n.BranchNode, visit)
	case *parse.PipeNode:
		if n == nil {
			return
		}
		visit(n)
		for _, cmd := range n.Cmds {
			for _, arg := range cmd.Args {
				switch a := arg.(type) {
				case *parse.PipeNode:
					walkPipes(a, visit)
				case *parse.ChainNode:
					walkPipes(a.Node, visit)
				}
			}
		}
	}
}

func walkBranch(branch *parse.BranchNode, visit func(*parse.PipeNode)) {
	walkPipes(branch.Pipe, visit)
	walkPipes(branch.List, visit)
	if branch.ElseList != nil {
		walkPipes(branch.ElseList, visit)
	}
}
