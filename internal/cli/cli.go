// Package cli implements the stream-only command interface.
package cli

import (
	"fmt"
	"html-ui/internal/catalog"
	"html-ui/internal/generate"
	"io"
	"strings"
)

const Version = "0.1.0"
const usage = `Usage: html-ui [--ts | --docs] <primitive>
       html-ui --list
       html-ui --help
       html-ui --version

Emit unstyled TypeScript primitive source to stdout (default or --ts).
--docs emits behavior, browser support, and styling instructions.
--list lists available primitive names. Flags may precede or follow the name.

Examples:
  html-ui accordion > Accordion.ts
  html-ui --docs accordion
  html-ui accordion | html-ui-vue > Accordion.vue

Framework converters are separate tools. No files are written by this CLI.
Use TypeScript's declaration emitter for .d.ts files; CEM is not used.
`

// Run returns 0 on success, 2 on invalid arguments, and 1 on output failure.
func Run(args []string, stdout, stderr io.Writer) int {
	mode, name := "", ""
	fail := func(message string) int { fmt.Fprintln(stderr, "html-ui: "+message); return 2 }
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			selected := ""
			switch arg {
			case "--ts":
				selected = "source"
			case "--docs":
				selected = "docs"
			case "--list":
				selected = "list"
			case "--help", "-h":
				selected = "help"
			case "--version":
				selected = "version"
			case "--manifest":
				return fail("CEM output was replaced by the TypeScript contract; use --ts")
			default:
				return fail(fmt.Sprintf("unknown flag %q; see --help", arg))
			}
			if mode != "" {
				return fail("select exactly one output mode")
			}
			mode = selected
		} else {
			if name != "" {
				return fail("expected one primitive name")
			}
			name = arg
		}
	}
	if mode == "" {
		mode = "source"
	}
	var output string
	switch mode {
	case "list", "help", "version":
		if name != "" {
			return fail("this mode does not accept a primitive name")
		}
		switch mode {
		case "list":
			var b strings.Builder
			for _, r := range catalog.All() {
				b.WriteString(r.Name + "\n")
			}
			output = b.String()
		case "help":
			output = usage
		case "version":
			output = "html-ui " + Version + "\n"
		}
	default:
		if name == "" {
			return fail("a primitive name is required; see --help")
		}
		r, ok := catalog.Find(name)
		if !ok {
			return fail(fmt.Sprintf("unknown primitive %q; use --list", name))
		}
		if mode == "docs" {
			output = generate.Docs(r)
		} else {
			output = generate.Source(r)
		}
	}
	n, err := io.WriteString(stdout, output)
	if err == nil && n != len(output) {
		err = io.ErrShortWrite
	}
	if err != nil {
		fmt.Fprintf(stderr, "html-ui: write output: %v\n", err)
		return 1
	}
	return 0
}
