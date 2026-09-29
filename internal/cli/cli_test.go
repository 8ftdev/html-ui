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

func TestContractVersionSelection(t *testing.T) {
	for _, args := range [][]string{
		{"accordion", "--contract-version=2"},
		{"--contract-version", "2", "--ts", "accordion"},
		{"--ts", "accordion", "--contract-version", "2"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, err bytes.Buffer
			if code := cli.Run(args, &out, &err); code != 0 || err.Len() != 0 {
				t.Fatalf("v2 selection failed (exit %d): %s", code, &err)
			}
			if !strings.Contains(out.String(), "export const contractVersion = 2 as const;") {
				t.Fatal("explicit v2 did not emit version 2")
			}
		})
	}
	var implicit, explicit, err bytes.Buffer
	cli.Run([]string{"accordion"}, &implicit, &err)
	if code := cli.Run([]string{"accordion", "--contract-version=2"}, &explicit, &err); code != 0 {
		t.Fatalf("explicit v2 failed: %s", &err)
	}
	if implicit.String() != explicit.String() {
		t.Fatal("default output must use v2")
	}
}

func TestInvalidContractVersionSelection(t *testing.T) {
	for _, args := range [][]string{
		{"accordion", "--contract-version=3"},
		{"accordion", "--contract-version=0"},
		{"accordion", "--contract-version="},
		{"accordion", "--contract-version=02"},
		{"accordion", "--contract-version"},
		{"accordion", "--contract-version", "--ts"},
		{"accordion", "--contract-version=2", "--contract-version=2"},
		{"accordion", "--contract-version=1", "--contract-version=2"},
		{"--docs", "accordion", "--contract-version=2"},
		{"--help", "--contract-version=2"},
		{"--list", "--contract-version=2"},
		{"--version", "--contract-version=2"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, err bytes.Buffer
			if code := cli.Run(args, &out, &err); code != 2 || out.Len() != 0 || err.Len() == 0 {
				t.Fatalf("exit %d, stdout=%q stderr=%q", code, out.String(), err.String())
			}
		})
	}
}
