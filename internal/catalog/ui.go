package catalog

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
)

//go:embed ui.json
var uiData []byte

// UIContract names owned styling targets independently of content slots.
type UIContract struct {
	Parts map[string]UIPart `json:"parts"`
}

type UIPart struct {
	Node      string                    `json:"node"`
	StyleRole string                    `json:"styleRole"`
	State     map[string]UIStateBinding `json:"state,omitempty"`
}

type UIStateBinding struct {
	Source UIStateSource `json:"source"`
}

// UIStateSource is either a pseudo-class or an attribute comparison on an
// explicit factory node. Pointer fields distinguish absence from false/empty.
type UIStateSource struct {
	Node      string  `json:"node"`
	Pseudo    string  `json:"pseudo,omitempty"`
	Attribute string  `json:"attribute,omitempty"`
	Present   *bool   `json:"present,omitempty"`
	Value     *string `json:"value,omitempty"`
}

func attachUI(recipes []Recipe) {
	var metadata map[string]UIContract
	decoder := json.NewDecoder(bytes.NewReader(uiData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		panic(fmt.Errorf("UI catalog: %w", err))
	}
	for i := range recipes {
		ui, ok := metadata[recipes[i].Name]
		if !ok {
			panic(fmt.Errorf("missing UI metadata for %s", recipes[i].Name))
		}
		recipes[i].UI = ui
		if err := validateUI(recipes[i]); err != nil {
			panic(fmt.Errorf("%s: %w", recipes[i].Name, err))
		}
		delete(metadata, recipes[i].Name)
	}
	if len(metadata) != 0 {
		panic("UI metadata contains unknown components")
	}
}

var uiName = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
var uiRole = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validateUI(r Recipe) error {
	nodes := map[string]Node{}
	for _, node := range r.Nodes {
		nodes[node.ID] = node
	}
	if root, ok := r.UI.Parts["root"]; !ok || root.Node != "root" {
		return fmt.Errorf("UI root part must target root")
	}
	seen := map[string]bool{}
	for name, part := range r.UI.Parts {
		if !uiName.MatchString(name) || !uiRole.MatchString(part.StyleRole) {
			return fmt.Errorf("invalid UI part name or style role: %s", name)
		}
		if _, ok := nodes[part.Node]; !ok || seen[part.Node] {
			return fmt.Errorf("UI part %s has unknown or duplicate node %s", name, part.Node)
		}
		seen[part.Node] = true
		for state, binding := range part.State {
			if !uiName.MatchString(state) {
				return fmt.Errorf("invalid UI state name %s", state)
			}
			if err := validateState(state, binding.Source, nodes); err != nil {
				return fmt.Errorf("UI part %s state %s: %w", name, state, err)
			}
		}
	}
	if len(seen) != len(nodes) {
		return fmt.Errorf("every owned node must have a UI part")
	}
	return nil
}

func validateState(state string, source UIStateSource, nodes map[string]Node) error {
	node, ok := nodes[source.Node]
	if !ok {
		return fmt.Errorf("unknown source node %s", source.Node)
	}
	if (source.Pseudo == "") == (source.Attribute == "") {
		return fmt.Errorf("expected exactly one pseudo or attribute source")
	}
	if source.Pseudo != "" {
		if source.Present != nil || source.Value != nil {
			return fmt.Errorf("pseudo source cannot have an attribute comparison")
		}
		switch source.Pseudo {
		case "hover", "active", "focus-visible", "focus-within":
		case "disabled":
			switch node.Tag {
			case "button", "input", "select", "textarea", "fieldset", "option", "optgroup":
			default:
				return fmt.Errorf("%s cannot be natively disabled", node.Tag)
			}
		case "checked":
			if node.Tag != "input" || (node.Attributes["type"] != "checkbox" && node.Attributes["type"] != "radio") {
				return fmt.Errorf("checked requires a checkbox or radio input")
			}
		case "required", "invalid":
			if node.Tag != "input" && node.Tag != "select" && node.Tag != "textarea" {
				return fmt.Errorf("%s requires a form control", source.Pseudo)
			}
			if node.Tag == "input" {
				switch node.Attributes["type"] {
				case "range", "hidden", "button", "submit", "reset", "image", "color":
					return fmt.Errorf("%s is unsupported for this input type", source.Pseudo)
				}
			}
		case "indeterminate":
			if node.Tag != "progress" && !(node.Tag == "input" && (node.Attributes["type"] == "checkbox" || node.Attributes["type"] == "radio")) {
				return fmt.Errorf("indeterminate requires progress, checkbox or radio")
			}
		case "popover-open":
			if _, ok := node.Attributes["popover"]; !ok {
				return fmt.Errorf("popover-open requires a popover target")
			}
		default:
			return fmt.Errorf("unsupported pseudo %s", source.Pseudo)
		}
	} else {
		if !uiRole.MatchString(source.Attribute) || (source.Present == nil) == (source.Value == nil) {
			return fmt.Errorf("expected a valid attribute and exactly one presence or value comparison")
		}
		if source.Attribute == "open" && node.Tag != "details" && node.Tag != "dialog" {
			return fmt.Errorf("open attribute requires details or dialog")
		}
	}
	if state == "disabled" && source.Pseudo != "disabled" && !(source.Attribute == "aria-disabled" && source.Value != nil && *source.Value == "true") {
		return fmt.Errorf("disabled requires native :disabled or aria-disabled=true")
	}
	return nil
}
