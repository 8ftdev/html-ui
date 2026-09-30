package catalog_test

import (
	"html-ui/internal/catalog"
	"reflect"
	"testing"
)

func TestNativeControlsPermitExternalAccessibleNames(t *testing.T) {
	for _, name := range []string{"input", "textarea", "checkbox", "radio", "switch", "slider", "select"} {
		t.Run(name, func(t *testing.T) {
			r, _ := catalog.Find(name)
			if r.Nodes[0].Tag != "label" || r.Nodes[1].ID != "control" {
				t.Fatal("native label root and control identity must stay stable")
			}
			for _, slot := range r.Slots {
				if slot.Name == "label" && slot.Required {
					t.Error("external accessible names must allow omitting the label slot")
				}
			}
			for prop, attr := range map[string]string{"id": "id", "ariaLabel": "aria-label", "ariaLabelledby": "aria-labelledby", "ariaDescribedby": "aria-describedby", "ariaInvalid": "aria-invalid"} {
				foundProp, foundBinding := false, false
				for _, p := range r.Props {
					if p.Name == prop {
						foundProp = !p.Required && len(p.Default) == 0
						if prop == "ariaInvalid" && p.Type != `"true" | "false"` {
							t.Error("aria-invalid must serialize explicit true/false values")
						}
					}
				}
				for _, b := range r.Bindings {
					if b.Prop == prop && b.Name == attr && b.Node == "control" && b.Kind == "attribute" {
						foundBinding = true
					}
				}
				if !foundProp || !foundBinding {
					t.Errorf("missing optional %s binding on actual control", prop)
				}
			}
		})
	}
}

func TestNativePresentationPartsKeepControlOwnedState(t *testing.T) {
	for _, tc := range []struct {
		name         string
		rootChildren []string
		parts        []string
	}{
		{"checkbox", []string{"control", "indicator", "label"}, []string{"indicator"}},
		{"radio", []string{"control", "indicator", "label"}, []string{"indicator"}},
		{"switch", []string{"control", "track", "label"}, []string{"track", "thumb"}},
		{"select", []string{"label", "control", "chevron"}, []string{"chevron"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := catalog.Find(tc.name)
			var rootChildren []string
			for _, c := range r.Children {
				if c.Parent == "root" {
					if c.Node != "" {
						rootChildren = append(rootChildren, c.Node)
					} else {
						rootChildren = append(rootChildren, c.Slot)
					}
				}
			}
			if !reflect.DeepEqual(rootChildren, tc.rootChildren) {
				t.Errorf("presentation anatomy order = %v, want %v", rootChildren, tc.rootChildren)
			}
			for _, name := range tc.parts {
				part, ok := r.UI.Parts[name]
				if !ok {
					t.Errorf("missing %s owned style target", name)
					continue
				}
				for _, n := range r.Nodes {
					if n.ID == part.Node && (n.Tag != "span" || n.Attributes["aria-hidden"] != "true") {
						t.Error("decorative part must be an aria-hidden HTML span")
					}
				}
				states := []string{"disabled", "focusVisible"}
				if tc.name != "select" {
					states = append(states, "checked")
				}
				for _, state := range states {
					if part.State[state].Source.Node != "control" {
						t.Errorf("%s.%s must reflect the native control", name, state)
					}
				}
			}
			if tc.name == "switch" {
				found := false
				for _, child := range r.Children {
					if child.Parent == "track" && child.Node == "thumb" {
						found = true
					}
				}
				if !found {
					t.Error("switch thumb must be inside its track")
				}
			}
			if err := catalog.Validate(r); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckboxIndeterminateUsesNativePropertyAndPseudo(t *testing.T) {
	r, _ := catalog.Find("checkbox")
	property := false
	for _, binding := range r.Bindings {
		if binding.Prop == "indeterminate" && binding.Node == "control" && binding.Kind == "property" && binding.Name == "indeterminate" {
			property = true
		}
	}
	if !property {
		t.Error("indeterminate must initialize the native property")
	}
	for _, part := range []string{"control", "indicator"} {
		source := r.UI.Parts[part].State["indeterminate"].Source
		if source.Node != "control" || source.Pseudo != "indeterminate" {
			t.Errorf("%s must read native :indeterminate", part)
		}
	}
}
