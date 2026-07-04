#!/bin/sh
# Integration checks for yup-echo, run inside a Debian (GNU coreutils) container.
#
# The shell builtin `echo` differs from coreutils, so the reference is the
# coreutils binary at /bin/echo, never the sh builtin.
#
# parity CASE  — yup-echo must produce byte-identical output to GNU /bin/echo.
# assert WANT  — yup-echo must produce WANT exactly (used where yup-echo
#                diverges from GNU by design; see cmd-echo COMPATIBILITY.md).
set -eu

ref=/bin/echo
fails=0

parity() {
	ours=$(yup-echo "$@" 2>/dev/null || true)
	gnu=$("$ref" "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  echo %s\n' "$*"
	else
		printf 'FAIL  parity  echo %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(yup-echo "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  echo %s\n' "$*"
	else
		printf 'FAIL  assert  echo %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Plain operands: joined by single spaces, trailing newline.
parity hello
parity hello world
parity

# -n: suppress the trailing newline.
parity -n hello
parity -n hello world
parity -n

# -e: interpret backslash escapes (\n and \t are the required cases).
parity -e 'a\tb'
parity -e 'a\nb'
parity -e 'tab\there'
parity -e 'line1\nline2'
parity -e 'a\tb\nc'

# -e combined with -n.
parity -n -e 'a\tb'

# -E: disable escape interpretation (the default); sequences stay literal.
parity -E 'a\tb'
parity 'a\tb'

# \c truncates remaining output and the trailing newline (matches GNU).
parity -e 'keep\cdrop'

# Documented divergence: \0NNN / \xHH numeric escapes are not interpreted,
# they remain literal (GNU /bin/echo would emit the encoded byte).
assert '\x41' -e '\x41'

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
