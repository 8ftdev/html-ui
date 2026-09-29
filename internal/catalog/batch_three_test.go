package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestFormAndContentBatch(t *testing.T) {
	for name, tag := range map[string]string{"empty": "div", "item": "div", "input-group": "div", "kbd": "kbd", "pagination": "nav", "typography": "div", "direction": "div", "radio-group": "fieldset", "radio": "label", "slider": "label"} {
		t.Run(name, func(t *testing.T) {
			r, ok := catalog.Find(name)
			if !ok {
				t.Fatalf("missing %s", name)
			}
			if err := catalog.Validate(r); err != nil {
				t.Fatal(err)
			}
			if r.Status != "native" || r.Nodes[0].Tag != tag {
				t.Fatal("unexpected native anatomy")
			}
		})
	}
}
func TestFormAndContentSemantics(t *testing.T) {
	r, ok := catalog.Find("radio")
	if !ok {
		t.Fatal("missing radio")
	}
	if r.Nodes[1].Attributes["type"] != "radio" {
		t.Fatal("radio must use native type")
	}
	for _, e := range r.Events {
		if len(e.State) > 0 {
			t.Fatal("radio selection is group-owned, not per-radio checked model")
		}
	}
	found := false
	for _, b := range r.Bindings {
		if b.Name == "defaultChecked" {
			found = true
		}
	}
	if !found {
		t.Fatal("radio needs native reset baseline")
	}
	g, ok := catalog.Find("input-group")
	if !ok {
		t.Fatal("missing input-group")
	}
	for _, n := range g.Nodes {
		if n.ID == "root" && n.Tag == "label" {
			t.Fatal("interactive addons must not nest inside label")
		}
	}
	found = false
	for _, b := range g.Bindings {
		if b.Node == "label" && b.Name == "for" && b.Prop == "id" {
			found = true
		}
	}
	if !found {
		t.Fatal("input-group label must associate by id")
	}
	s, _ := catalog.Find("slider")
	found = false
	for _, e := range s.Events {
		if e.State["value"] == "valueAsNumber" {
			found = true
		}
	}
	if !found {
		t.Fatal("slider must expose numeric native value")
	}
}
