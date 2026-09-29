package generate

import (
	"encoding/json"
	"fmt"
	"html-ui/internal/catalog"
	"strings"
)

func uiSource(r catalog.Recipe, name string) string {
	var b strings.Builder
	p := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	metadata := struct {
		Component string `json:"component"`
		Behavior  struct {
			Kind         string   `json:"kind"`
			Requirements []string `json:"requirements"`
		} `json:"behavior"`
		Parts map[string]catalog.UIPart `json:"parts"`
	}{Component: r.Name, Parts: r.UI.Parts}
	metadata.Behavior.Kind = r.Status
	metadata.Behavior.Requirements = r.Requirements
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		panic(err)
	} // This closed literal model contains no unsupported JSON values.
	p("/** Portable anatomy and state sources. Preserve through framework and theme transforms. */\nexport const ui = %s as const;\n\n", data)
	p("/** Styling operations only; adapters implement composition without changing native behavior. */\nexport type %sStyle<Style = string> = Style | { mode: \"replace\"; value: Style } | { mode: \"omit\" };\n\n", name)
	p("/** Owned styling targets only. Content slots are not implicitly wrapped. */\nexport interface %sClasses<Style = string> {\n", name)
	for _, partName := range keys(r.UI.Parts) {
		part := r.UI.Parts[partName]
		p("  %s?: {\n    base?: %sStyle<Style>;\n    unstyled?: boolean;\n", partName, name)
		if len(part.State) > 0 {
			p("    state?: {\n")
			for _, state := range keys(part.State) {
				p("      %s?: %sStyle<Style>;\n", state, name)
			}
			p("    };\n")
		}
		p("  };\n")
	}
	p("}\n\n")
	return b.String()
}
