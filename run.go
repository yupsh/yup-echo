package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	command "github.com/gloo-foo/cmd-echo"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "echo"

const (
	flagNoNewline  = "no-newline"
	flagEscapes    = "escapes"
	flagNoEscapes  = "no-escapes"
	escapeTruncate = `\c`
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `echo [OPTIONS] [STRING...]

echo the STRING(s) to standard output, separated by single spaces.`

// escaper expands the GNU echo backslash escapes that -e enables. \0NNN and
// \xHH are intentionally unsupported (see cmd-echo COMPATIBILITY.md); \c is
// handled separately because it truncates the remaining output.
var escaper = strings.NewReplacer(
	`\\`, "\\",
	`\a`, "\a",
	`\b`, "\b",
	`\e`, "\x1b",
	`\f`, "\f",
	`\n`, "\n",
	`\r`, "\r",
	`\t`, "\t",
	`\v`, "\v",
)

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the echo CLI against the injected version and I/O,
// returning the process exit code. echo does not read stdin or the filesystem;
// both are injected for a uniform, testable wiring shape.
func run(version string, args []string, _ io.Reader, stdout, stderr io.Writer, _ afero.Fs) int {
	cmd := newCommand(version, stdout)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "display a line of text",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: flagNoNewline, Aliases: []string{"n"}, Usage: "do not output the trailing newline"},
			&cli.BoolFlag{
				Name:    flagEscapes,
				Aliases: []string{"e"},
				Usage:   "enable interpretation of backslash escapes",
			},
			&cli.BoolFlag{
				Name:    flagNoEscapes,
				Aliases: []string{"E"},
				Usage:   "disable interpretation of backslash escapes (default)",
			},
		},
		Action: action(stdout),
	}
}

func action(stdout io.Writer) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		text, truncated := render(cmd, joined(ctx, cmd.Args().Slice()))
		return emit(stdout, text, cmd.Bool(flagNoNewline) || truncated)
	}
}

// joined drives cmd-echo to assemble the single space-separated line, the one
// behavior the wrapper delegates to the command rather than reimplementing.
// command.Echo streams exactly one synthetic in-memory item, so its Collect
// cannot fail; the error is discarded as structurally impossible.
func joined(ctx context.Context, args []string) string {
	items, _ := gloo.Collect(ctx, command.Echo(args...).Stream(ctx))
	return string(items[0])
}

// render applies backslash-escape interpretation when -e is set and -E is not,
// reporting whether a \c escape truncated the output (which also drops the
// trailing newline).
func render(cmd *cli.Command, text string) (string, bool) {
	if cmd.Bool(flagEscapes) && !cmd.Bool(flagNoEscapes) {
		return escape(text)
	}
	return text, false
}

// escape expands the supported backslash escapes, honoring \c which truncates
// all remaining output (including the trailing newline). The bool reports a \c
// truncation so the caller can suppress the newline.
func escape(text string) (string, bool) {
	truncated := false
	if i := strings.Index(text, escapeTruncate); i >= 0 {
		text, truncated = text[:i], true
	}
	return escaper.Replace(text), truncated
}

// emit writes text to stdout, appending a newline unless it was suppressed by
// -n or a \c escape.
func emit(stdout io.Writer, text string, noNewline bool) error {
	if noNewline {
		_, err := io.WriteString(stdout, text)
		return err
	}
	return gloo.OutputString(stdout, text)
}
