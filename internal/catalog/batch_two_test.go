package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestCompositionBatchNativeContracts(t *testing.T) {
	roots := map[string]string{"alert": "div", "badge": "span", "aspect-ratio": "div", "breadcrumb": "nav", "button-group": "div", "label": "label", "skeleton": "div", "spinner": "span", "table": "div", "textarea": "label"}
	for name, tag := range roots {
		t.Run(name, func(t *testing.T) {
			r, ok := catalog.Find(name)
			if !ok {
				t.Fatalf("missing native contract %s", name)
			}
			if err := catalog.Validate(r); err != nil {
				t.Fatal(err)
			}
			if r.Status != "native" || r.Nodes[0].Tag != tag {
				t.Fatalf("unexpected native root for %s", name)
			}
		})
	}
}

func TestCompositionBatchSemantics(t *testing.T) {
	for _, tc := range []struct{ component, node, attribute, value string }{
		{"alert", "root", "role", "alert"}, {"alert", "icon", "aria-hidden", "true"},
		{"button-group", "root", "role", "group"}, {"skeleton", "root", "aria-hidden", "true"},
		{"spinner", "root", "role", "status"}, {"spinner", "indicator", "aria-hidden", "true"},
	} {
		r, ok := catalog.Find(tc.component)
		if !ok {
			t.Errorf("missing %s", tc.component)
			continue
		}
		found := false
		for _, n := range r.Nodes {
			if n.ID == tc.node && n.Attributes[tc.attribute] == tc.value {
				found = true
			}
		}
		if !found {
			t.Errorf("%s lacks %s.%s=%s", tc.component, tc.node, tc.attribute, tc.value)
		}
	}
	r, ok := catalog.Find("textarea")
	if !ok {
		t.Fatal("missing textarea")
	}
	bindings := map[string]bool{}
	for _, b := range r.Bindings {
		if b.Node == "control" {
			bindings[b.Name] = true
		}
	}
	for _, name := range []string{"defaultValue", "value", "disabled", "required", "readOnly", "rows", "cols", "minlength", "maxlength"} {
		if !bindings[name] {
			t.Errorf("textarea lacks %s", name)
		}
	}
	for _, name := range []string{"disabled", "focusVisible", "invalid", "required", "readOnly"} {
		if _, ok := r.UI.Parts["control"].State[name]; !ok {
			t.Errorf("textarea lacks %s state", name)
		}
	}
}
