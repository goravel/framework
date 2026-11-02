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
	// Args are the string constants passed to the function, in order; nil when Dynamic.
	Args []string
	// Dynamic reports that an argument is computed at render time, so the call
	// cannot be resolved by reading the template.
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
				if len(args) == 0 && i > 0 {
					args = pipe.Cmds[i-1].Args
				}

				call := Call{Template: t.Name(), Args: make([]string, 0, len(args))}
				for _, arg := range args {
					str, ok := arg.(*parse.StringNode)
					if !ok {
						call = Call{Template: t.Name(), Dynamic: true}
						break
					}
					call.Args = append(call.Args, str.Text)
				}

				calls = append(calls, call)
			}
		})
	}

	return calls
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
