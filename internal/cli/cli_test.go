package cli_test

import (
	"bytes"
	"errors"
	"html-ui/internal/cli"
	"strings"
	"testing"
)

func TestModes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"accordion"}, "export interface AccordionProps"},
		{[]string{"--ts", "accordion"}, "export interface AccordionProps"},
		{[]string{"accordion", "--ts"}, "export interface AccordionProps"},
		{[]string{"--docs", "accordion"}, "# Accordion"},
		{[]string{"accordion", "--docs"}, "# Accordion"},
		{[]string{"--list"}, "accordion\nalert-dialog\n"},
		{[]string{"--help"}, "Usage:"}, {[]string{"-h"}, "Usage:"},
		{[]string{"--version"}, "html-ui "},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, err bytes.Buffer
			if code := cli.Run(tc.args, &out, &err); code != 0 {
				t.Fatalf("exit %d: %s", code, err.String())
			}
			if err.Len() != 0 || !strings.Contains(out.String(), tc.want) {
				t.Fatalf("stdout=%q stderr=%q", out.String(), err.String())
			}
		})
	}
}
func TestInvalidArgumentsLeaveStdoutEmpty(t *testing.T) {
	for _, args := range [][]string{nil, {"missing"}, {"accordion", "button"}, {"--docs"}, {"--ts"}, {"--wat"}, {"--ts", "--docs", "accordion"}, {"--list", "accordion"}, {"--help", "accordion"}, {"--list", "--docs"}, {"--manifest", "accordion"}, {"--version", "--list"}, {"--ts", "--ts", "accordion"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, err bytes.Buffer
			if cli.Run(args, &out, &err) == 0 || out.Len() != 0 || err.Len() == 0 {
				t.Fatalf("stdout=%q stderr=%q", out.String(), err.String())
			}
		})
	}
}

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) { return 0, errors.New("output unavailable") }
func TestWriteFailure(t *testing.T) {
	var err bytes.Buffer
	if cli.Run([]string{"accordion"}, brokenWriter{}, &err) != 1 || !strings.Contains(err.String(), "output unavailable") {
		t.Fatalf("not reported: %s", err.String())
	}
}
func TestSourceAlias(t *testing.T) {
	var a, b, e bytes.Buffer
	cli.Run([]string{"accordion"}, &a, &e)
	cli.Run([]string{"--ts", "accordion"}, &b, &e)
	if a.String() != b.String() {
		t.Fatal("--ts differs from default")
	}
}
