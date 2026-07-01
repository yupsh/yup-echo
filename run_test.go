package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "single argument",
			args:    []string{"echo", "hello"},
			wantOut: "hello\n",
		},
		{
			name:    "multiple arguments joined by spaces",
			args:    []string{"echo", "hello", "world"},
			wantOut: "hello world\n",
		},
		{
			name:    "no arguments emits a blank line",
			args:    []string{"echo"},
			wantOut: "\n",
		},
		{
			name:    "no-newline suppresses the trailing newline",
			args:    []string{"echo", "-n", "hello"},
			wantOut: "hello",
		},
		{
			name:    "no-newline with no arguments emits nothing",
			args:    []string{"echo", "-n"},
			wantOut: "",
		},
		{
			name:    "escapes interprets newline and tab",
			args:    []string{"echo", "-e", `a\tb\nc`},
			wantOut: "a\tb\nc\n",
		},
		{
			name:    "escapes truncates at backslash-c",
			args:    []string{"echo", "-e", `keep\cdrop`},
			wantOut: "keep",
		},
		{
			name:    "no-escapes leaves backslash sequences literal",
			args:    []string{"echo", "-E", `a\tb`},
			wantOut: `a\tb` + "\n",
		},
		{
			name:    "no-escapes overrides escapes",
			args:    []string{"echo", "-e", "-E", `a\tb`},
			wantOut: `a\tb` + "\n",
		},
		{
			name:    "escapes and no-newline combine",
			args:    []string{"echo", "-n", "-e", `a\tb`},
			wantOut: "a\tb",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"echo", "--version"},
			wantOut: "echo version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"echo", "--nope"},
			wantCode:   1,
			wantErrSub: "echo:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(""), &out, &errOut, afero.NewMemMapFs())

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

// errWriter fails every write, exercising emit's error path on both the -n
// (io.WriteString) and default (gloo.OutputString) branches.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestRun_WriteError(t *testing.T) {
	for _, args := range [][]string{
		{"echo", "hello"},
		{"echo", "-n", "hello"},
	} {
		var errOut bytes.Buffer
		code := run("", args, strings.NewReader(""), errWriter{}, &errOut, afero.NewMemMapFs())
		if code != 1 {
			t.Fatalf("args %v: exit code = %d, want 1", args, code)
		}
		if !strings.Contains(errOut.String(), "echo:") {
			t.Fatalf("args %v: stderr = %q, want substring %q", args, errOut.String(), "echo:")
		}
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
