package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestCalendarGridContract(t *testing.T) {
	r, ok := catalog.Find("date-grid")
	if !ok {
		t.Fatal("missing portable date-grid primitive")
	}
	for _, part := range []string{"root", "label", "control", "trigger", "value", "icon", "popup", "header", "previous", "caption", "next", "grid", "weekdays", "weekday", "body", "week", "cell", "day"} {
		if r.UI.Parts[part].Node == "" {
			t.Errorf("missing owned %s", part)
		}
	}
	var dateProxy, model bool
	for _, n := range r.Nodes {
		if n.ID == "control" && n.Tag == "input" && n.Attributes["type"] == "date" && n.Attributes["tabindex"] == "-1" {
			dateProxy = true
		}
	}
	for _, e := range r.Events {
		if e.Name == "input" && e.Target == "control" && e.State["value"] == "value" {
			model = true
		}
	}
	if !dateProxy || !model {
		t.Fatal("calendar must preserve the native date value/form model")
	}
}
