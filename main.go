// Command yup-echo is the CLI wrapper around github.com/gloo-foo/cmd-echo.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-echo"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const name = "echo"

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `echo [STRING...]

echo the STRING(s) to standard output, separated by single spaces.`

// spec declares the echo wrapper. echo is a source command: it produces its
// line directly, so build returns it as the whole pipeline (a nil filter).
var spec = clix.Spec{
	Name:     name,
	Summary:  "display a line of text",
	Synopsis: synopsis,
	Build:    build,
}

// build maps the invocation to echo's pipeline: the STRING operands become the
// echo source, with no filter. Each operand is converted to the command's typed
// Text argument.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	operands := inv.Args.Args().Slice()
	words := make([]command.Text, len(operands))
	for i, o := range operands {
		words[i] = command.Text(o)
	}
	return command.Echo(words...), nil, nil
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
