// Package generate renders the public TypeScript convention and its documentation.
package generate

import (
	"fmt"
	"html-ui/internal/catalog"
	"sort"
	"strconv"
	"strings"
)

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func title(name string) string {
	var out string
	for _, p := range strings.Split(name, "-") {
		out += strings.ToUpper(p[:1]) + p[1:]
	}
	return out
}
func optional(required bool) string {
	if required {
		return ""
	}
	return "?"
}
func comment(s string) string { return strings.ReplaceAll(s, "*/", "* /") }

// Source emits a standalone TypeScript module. No framework or runtime imports are needed.
func Source(r catalog.Recipe) string {
	var b strings.Builder
	p := func(f string, a ...any) { fmt.Fprintf(&b, f, a...) }
	t := title(r.Name)
	factory := strings.ToLower(t[:1]) + t[1:]
	if factory == "switch" {
		factory = "switchControl"
	}
	p("/**\n * %s\n * %s\n * Behavior: %s. See `html-ui --docs %s`.\n * Source conversion contract v1; native factory initializes once.\n */\n", r.Title, comment(r.Summary), r.Status, r.Name)
	p("export const contractVersion = 1 as const;\n\nexport interface %sProps {\n", t)
	for _, v := range r.Props {
		p("  /** %s */\n  %s%s: %s;\n", comment(v.Description), v.Name, optional(v.Required), v.Type)
	}
	p("}\n\nexport const defaults = {\n")
	for _, v := range r.Props {
		if len(v.Default) > 0 {
			p("  %s: %s,\n", v.Name, v.Default)
		}
	}
	p("} as const satisfies Partial<%sProps>;\n\nexport interface %sSlots<Content = HTMLElement> {\n", t, t)
	for _, s := range r.Slots {
		scope := ""
		if len(s.Scope) > 0 {
			var pairs []string
			for _, v := range s.Scope {
				pairs = append(pairs, v.Name+": "+v.Type)
			}
			scope = "scope: { " + strings.Join(pairs, "; ") + " }"
		}
		p("  /** %s */\n  %s%s: (%s) => Content;\n", comment(s.Description), s.Name, optional(s.Required), scope)
	}
	p("}\n\nexport interface %sEvents {\n", t)
	for _, e := range r.Events {
		p("  /** %s */\n  %s: %s;\n", comment(e.Description), e.Name, e.Type)
	}
	p("}\n\n/** Native event targets and DOM properties to read after an event. No listeners are installed. */\nexport const nativeEvents = {\n")
	for _, e := range r.Events {
		p("  %s: { target: %q, state: {", e.Name, e.Target)
		for _, k := range keys(e.State) {
			p(" %s: %q,", k, e.State[k])
		}
		p(" } },\n")
	}
	p("} as const;\n\n/** Build initial native structure; converters parse this body without executing it. */\nexport function %s(props: %sProps, slots: %sSlots<HTMLElement>): %s {\n", factory, t, t, r.RootType)
	if len(r.Props) > 0 {
		var fields []string
		for _, v := range r.Props {
			f := v.Name
			if len(v.Default) > 0 {
				f += " = defaults." + v.Name
			}
			fields = append(fields, f)
		}
		p("  const { %s } = props;\n", strings.Join(fields, ", "))
	}
	for _, n := range r.Nodes {
		p("  const %s = document.createElement(%q);\n", n.ID, n.Tag)
	}
	for _, n := range r.Nodes {
		for _, k := range keys(n.Attributes) {
			p("  %s.setAttribute(%q, %q);\n", n.ID, k, n.Attributes[k])
		}
	}
	props := map[string]catalog.Prop{}
	for _, v := range r.Props {
		props[v.Name] = v
	}
	for _, v := range r.Bindings {
		prefix := ""
		prop := props[v.Prop]
		if !prop.Required && len(prop.Default) == 0 {
			prefix = fmt.Sprintf("if (%s !== undefined) ", v.Prop)
		}
		if v.Kind == "property" {
			p("  %s%s.%s = %s;\n", prefix, v.Node, v.Name, v.Prop)
		} else {
			value := v.Prop
			if prop.Type != "string" {
				value = "String(" + value + ")"
			}
			p("  %s%s.setAttribute(%q, %s);\n", prefix, v.Node, v.Name, value)
		}
	}
	slots := map[string]catalog.Slot{}
	for _, s := range r.Slots {
		slots[s.Name] = s
	}
	for _, c := range r.Children {
		if c.Node != "" {
			p("  %s.append(%s);\n", c.Parent, c.Node)
			continue
		}
		s := slots[c.Slot]
		prefix := ""
		if !s.Required {
			prefix = "if (slots." + s.Name + " !== undefined) "
		}
		scope := ""
		if len(s.Scope) > 0 {
			var fields []string
			for _, v := range s.Scope {
				field := v.Name
				if v.Name != v.Prop {
					field += ": " + v.Prop
				}
				fields = append(fields, field)
			}
			scope = "{ " + strings.Join(fields, ", ") + " }"
		}
		p("  %s%s.append(slots.%s(%s));\n", prefix, c.Parent, s.Name, scope)
	}
	p("  return root;\n}\n")
	return b.String()
}

// Docs renders behavior and styling instructions separately from source.
func Docs(r catalog.Recipe) string {
	var b strings.Builder
	p := func(f string, a ...any) { fmt.Fprintf(&b, f, a...) }
	p("# %s\n\n%s\n\n## Behavior\n\nClassification: **%s**. Native factories initialize once; converters own subsequent rendering and state.\n\n", r.Title, r.Summary, r.Status)
	for _, s := range r.Requirements {
		p("- %s\n", s)
	}
	p("\n## Props\n\n| Name | Type | Required | Default | Meaning |\n| --- | --- | --- | --- | --- |\n")
	for _, v := range r.Props {
		d := "—"
		if len(v.Default) > 0 {
			d = string(v.Default)
		}
		p("| %s | `%s` | %t | `%s` | %s |\n", v.Name, strings.ReplaceAll(v.Type, "|", "\\|"), v.Required, d, v.Description)
	}
	p("\n## Slots\n\nSlots are render signatures with a replaceable content type; native calls expect HTMLElement. No Shadow DOM projection is implied.\n\n")
	for _, s := range r.Slots {
		p("- `%s` (required: %t): %s", s.Name, s.Required, s.Description)
		for _, v := range s.Scope {
			p(" Scope `%s: %s` comes from `%s`.", v.Name, v.Type, v.Prop)
		}
		p("\n")
	}
	p("\n## Events\n\nThe `nativeEvents` constant identifies factory-local targets and state properties to read. It installs no listeners. Converters decide between props, local state, and two-way models.\n\n")
	for _, e := range r.Events {
		p("- `%s` (`%s`) on `%s`: %s", e.Name, e.Type, e.Target, e.Description)
		for _, k := range keys(e.State) {
			p(" Read `%s.%s` into `%s`.", e.Target, e.State[k], k)
		}
		p("\n")
	}
	p("\n## Styling\n\nNo CSS or reset is emitted. Browser default styling remains.\n\n")
	for _, s := range r.Styling {
		p("- %s\n", s)
	}
	if len(r.Features) > 0 {
		p("\n## Browser features\n\nReviewed snapshot, not a live compatibility guarantee. Test the target browsers.\n\n")
		for _, f := range r.Features {
			p("- [%s](https://web-platform-dx.github.io/web-features-explorer/features/%s/): %s (reviewed %s).\n", f.ID, f.ID, f.Status, f.Date)
		}
	}
	p("\n## Conversion\n\nUse `html-ui %s` for TypeScript source. Parse interfaces, literal defaults, `nativeEvents`, and the constrained factory AST. Never run incoming source to discover its structure. Omit undefined optional attributes; use DOM boolean properties or attribute presence/absence, never the string %s. Translate slot outlets to the target renderer. CSS, interaction listeners, focus management, and framework code belong to the converter.\n", r.Name, strconv.Quote("false"))
	return b.String()
}
