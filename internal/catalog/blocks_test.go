package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestLayoutBlocks(t *testing.T) {
	for _, name := range []string{"card", "grid", "icon"} {
		r, ok := catalog.Find(name)
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		if err := catalog.Validate(r); err != nil {
			t.Fatal(err)
		}
		if len(r.Slots) != 1 || r.Slots[0].Name != "default" {
			t.Errorf("%s needs content slot", name)
		}
	}
}
func TestInputControlAttributes(t *testing.T) {
	r, _ := catalog.Find("input")
	for _, name := range []string{"id", "type", "autocomplete"} {
		found := false
		for _, b := range r.Bindings {
			if b.Prop == name && b.Node == "control" {
				found = true
			}
		}
		if !found {
			t.Errorf("missing control binding %s", name)
		}
	}
}
func TestButtonAccessibleName(t *testing.T) {
	r, _ := catalog.Find("button")
	for _, b := range r.Bindings {
		if b.Prop == "ariaLabel" && b.Name == "aria-label" && b.Node == "root" {
			return
		}
	}
	t.Fatal("missing accessible name binding")
}
