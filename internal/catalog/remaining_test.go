package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestRemainingNativeBuildingBlocks(t *testing.T) {
	for _, name := range []string{"attachment", "bubble", "date-field", "carousel", "chart", "marker", "message", "message-scroller", "questionnaire", "resizable", "sidebar", "notification"} {
		t.Run(name, func(t *testing.T) {
			r, ok := catalog.Find(name)
			if !ok {
				t.Fatalf("missing %s", name)
			}
			if err := catalog.Validate(r); err != nil {
				t.Fatal(err)
			}
			if r.Status != "native" {
				t.Fatal("native block requires an adapter")
			}
		})
	}
}
