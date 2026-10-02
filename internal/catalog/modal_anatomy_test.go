package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestModalOwnedAnatomy(t *testing.T) {
	for _, name := range []string{"dialog", "alert-dialog", "drawer"} {
		t.Run(name, func(t *testing.T) {
			r, _ := catalog.Find(name)
			for _, part := range []string{"header", "description", "content", "footer"} {
				if r.UI.Parts[part].Node != part {
					t.Errorf("missing owned %s", part)
				}
			}
			descriptionID, descriptionSlot, footerSlot := false, false, false
			for _, p := range r.Props {
				if p.Name == "descriptionId" {
					descriptionID = true
				}
			}
			for _, s := range r.Slots {
				if s.Name == "description" {
					descriptionSlot = true
				}
				if s.Name == "footer" {
					footerSlot = true
				}
			}
			if !descriptionID || !descriptionSlot || !footerSlot {
				t.Fatal("missing connected description/footer contract")
			}
			for _, b := range r.Bindings {
				if b.Node == "description" && b.Name == "id" && b.Prop == "descriptionId" {
					return
				}
			}
			t.Fatal("description must own the described element ID")
		})
	}
}
