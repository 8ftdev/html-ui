package generate_test

import (
	"html-ui/internal/catalog"
	"html-ui/internal/generate"
	"strings"
	"testing"
)

func TestCatalogCoverage(t *testing.T) {
	names := strings.Fields("empty item input-group kbd pagination typography direction radio accordion alert alert-dialog aspect-ratio badge breadcrumb button-group label skeleton spinner table textarea autocomplete avatar button card grid icon checkbox checkbox-group collapsible combobox context-menu dialog drawer field fieldset form input menu menubar meter navigation-menu number-field otp-field popover preview-card progress radio-group scroll-area select separator slider switch tabs toast toggle toggle-group toolbar tooltip")
	if len(catalog.All()) != len(names) {
		t.Fatalf("want %d primitives, got %d", len(names), len(catalog.All()))
	}
	for _, name := range names {
		r, ok := catalog.Find(name)
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		if r.Name != name {
			t.Errorf("wrong lookup: %s", r.Name)
		}
		if err := catalog.Validate(r); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
func TestAccordionContract(t *testing.T) {
	r, _ := catalog.Find("accordion")
	source := generate.Source(r)
	for _, want := range []string{"export interface AccordionProps", "name?: string;", "open?: boolean;", "open: false,", "satisfies Partial<AccordionProps>", "AccordionSlots<Content = HTMLElement>", "summary: () => Content;", "content?: (scope: { open: boolean }) => Content;", "toggle: ToggleEvent;", `target: "root"`, `open: "open"`, "HTMLDetailsElement", `document.createElement("details")`, `if (name !== undefined)`, `root.open = open;`, `summary.append(slots.summary());`, `root.append(summary);`, `content.append(slots.content({ open }));`} {
		if !strings.Contains(source, want) {
			t.Errorf("missing %q in:\n%s", want, source)
		}
	}
	if strings.Contains(source, `setAttribute("open"`) {
		t.Fatal("boolean open must use DOM property")
	}
	if generate.Source(r) != source {
		t.Fatal("nondeterministic output")
	}
}
func TestUnstyledAndDocumented(t *testing.T) {
	for _, r := range catalog.All() {
		t.Run(r.Name, func(t *testing.T) {
			source := generate.Source(r)
			docs := generate.Docs(r)
			for _, bad := range []string{"<style", ".style", "addEventListener", "customElements.define", "from \"vue\"", "from \"react\""} {
				if strings.Contains(source, bad) {
					t.Errorf("unexpected %s", bad)
				}
			}
			if !strings.Contains(docs, "Styling") || !strings.Contains(docs, "Behavior") {
				t.Fatal("missing guidance")
			}
			if len(r.Requirements) == 0 {
				t.Fatal("missing behavior limits")
			}
		})
	}
	for _, name := range []string{"tabs", "menu", "combobox", "tooltip", "toast"} {
		r, _ := catalog.Find(name)
		if r.Status != "adapter-required" || !strings.Contains(generate.Docs(r), "adapter-required") {
			t.Errorf("%s claims native completeness", name)
		}
	}
}
func TestProgressOmission(t *testing.T) {
	r, _ := catalog.Find("progress")
	s := generate.Source(r)
	if !strings.Contains(s, "if (value !== undefined)") {
		t.Fatal("indeterminate progress requires absent value")
	}
}
