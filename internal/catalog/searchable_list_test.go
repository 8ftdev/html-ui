package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestSearchableListsKeepNativeValuesSeparateFromEditableSearch(t *testing.T) {
	for _, name := range []string{"combobox-list", "command-list"} {
		t.Run(name, func(t *testing.T) {
			r, ok := catalog.Find(name)
			if !ok {
				t.Fatalf("missing searchable primitive %s", name)
			}
			if err := catalog.Validate(r); err != nil {
				t.Fatal(err)
			}
			if r.Status != "adapter-required" {
				t.Fatal("search requires an explicit local adapter")
			}
			nodes := map[string]catalog.Node{}
			for _, n := range r.Nodes {
				nodes[n.ID] = n
			}
			if nodes["input"].Tag != "input" || nodes["input"].Attributes["role"] != "combobox" || nodes["input"].Attributes["aria-autocomplete"] != "list" {
				t.Error("editable search needs list autocomplete semantics")
			}
			if nodes["control"].Tag != "select" || nodes["control"].Attributes["aria-hidden"] != "true" {
				t.Error("committed values need the native proxy")
			}
			if nodes["option"].Attributes["role"] != "option" {
				t.Error("missing owned repeated option")
			}
			if _, ok := nodes["option"].Attributes["hidden"]; !ok {
				t.Error("prototype must be hidden")
			}
			if nodes["popup"].Attributes["role"] != "listbox" {
				t.Error("choices require a listbox")
			}
			if name == "combobox-list" && nodes["popup"].Attributes["popover"] != "auto" {
				t.Error("combobox requires a native popup")
			}
			if name == "command-list" && nodes["popup"].Attributes["popover"] != "" {
				t.Error("Command list must remain inline")
			}
			for _, b := range r.Bindings {
				if b.Node == "input" && b.Name == "value" {
					t.Error("query must not write the committed model")
				}
				if b.Prop == "name" && b.Node != "control" {
					t.Error("search text must not be submitted")
				}
			}
			if r.UI.Parts["input"].State["expanded"].Source.Attribute != "aria-expanded" {
				t.Error("expanded state must follow visible input")
			}
		})
	}
}
