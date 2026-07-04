package main

import (
	"context"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  spec.Flags,
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestBuild_Source(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"none", []string{name}},
		{"words", []string{name, "hello", "world"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inv := clix.Invocation{Args: parse(t, tc.args...), Fs: afero.NewMemMapFs()}
			src, filter, err := build(inv)
			if err != nil || src == nil || filter != nil {
				t.Fatalf("build: src=%v filter=%v err=%v (want source, nil filter)", src, filter, err)
			}
		})
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}
