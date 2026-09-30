package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestFieldOwnsContentAndErrorPresentation(t *testing.T) {
	r, _ := catalog.Find("field")
	for _, name := range []string{"content", "error"} {
		if _, ok := r.UI.Parts[name]; !ok {
			t.Errorf("missing %s styling target", name)
		}
	}
	for _, state := range []string{"invalid", "presentationDisabled"} {
		attr := "data-invalid"
		if state == "presentationDisabled" {
			attr = "data-disabled"
		}
		source := r.UI.Parts["root"].State[state].Source
		if source.Node != "root" || source.Attribute != attr || source.Value == nil || *source.Value != "true" {
			t.Errorf("%s presentation state must compare its explicit true hook", state)
		}
		for _, part := range []string{"label", "content", "error"} {
			if r.UI.Parts[part].State[state].Source.Node != "root" {
				t.Errorf("%s %s presentation must follow the Field root", part, state)
			}
		}
	}
	errorID := false
	for _, b := range r.Bindings {
		if b.Node == "error" && b.Name == "id" && b.Prop == "errorId" {
			errorID = true
		}
	}
	if !errorID {
		t.Error("error text needs its caller-supplied description link target")
	}
	for _, node := range []string{"label", "description", "error"} {
		found := false
		for _, child := range r.Children {
			if child.Parent == "content" && child.Node == node {
				found = true
			}
		}
		if !found {
			t.Errorf("%s must be owned by the content presentation wrapper", node)
		}
	}
	if err := catalog.Validate(r); err != nil {
		t.Fatal(err)
	}
}

func TestFieldsetKeepsLegendBeforeDescriptionAndNativeContent(t *testing.T) {
	r, _ := catalog.Find("fieldset")
	if _, ok := r.UI.Parts["description"]; !ok {
		t.Error("missing fieldset description styling target")
	}
	var rootOrder []string
	for _, child := range r.Children {
		if child.Parent == "root" {
			if child.Node != "" {
				rootOrder = append(rootOrder, child.Node)
			} else {
				rootOrder = append(rootOrder, child.Slot)
			}
		}
	}
	if len(rootOrder) != 3 || rootOrder[0] != "legend" || rootOrder[1] != "description" || rootOrder[2] != "content" {
		t.Errorf("legend must remain first child ahead of description and controls: %v", rootOrder)
	}
	if err := catalog.Validate(r); err != nil {
		t.Fatal(err)
	}
}
