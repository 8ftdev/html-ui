package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestSelectListKeepsLabelsAndFormSemanticsOnTheirRespectiveControls(t *testing.T) {
	r, ok := catalog.Find("select-list")
	if !ok {
		t.Fatal("missing select-list")
	}
	if err := catalog.Validate(r); err != nil {
		t.Fatal(err)
	}
	if r.Status != "adapter-required" || r.RootType != "HTMLDivElement" {
		t.Fatal("custom selection needs a distinct adapter-owned div root")
	}
	nodes := map[string]catalog.Node{}
	for _, node := range r.Nodes {
		nodes[node.ID] = node
	}
	if nodes["control"].Tag != "select" || nodes["control"].Attributes["aria-hidden"] != "true" || nodes["control"].Attributes["tabindex"] != "-1" {
		t.Error("form submission must use a native select excluded from the accessibility tree and tab order")
	}
	if _, hidden := nodes["control"].Attributes["hidden"]; hidden {
		t.Error("native proxy must remain available for constraint validation")
	}
	if nodes["label"].Tag != "label" || nodes["trigger"].Tag != "button" || nodes["trigger"].Attributes["type"] != "button" || nodes["trigger"].Attributes["role"] != "combobox" {
		t.Error("labels must name a non-submitting combobox button")
	}
	if nodes["trigger"].Attributes["aria-haspopup"] != "listbox" || nodes["trigger"].Attributes["aria-expanded"] != "false" || nodes["popup"].Attributes["role"] != "listbox" || nodes["popup"].Attributes["popover"] != "auto" {
		t.Error("trigger and popup must establish the listbox relationship")
	}
	if nodes["option"].Tag != "div" || nodes["option"].Attributes["role"] != "option" || nodes["option"].Attributes["aria-selected"] != "false" {
		t.Error("projected choices need an owned option prototype")
	}
	if _, hidden := nodes["option"].Attributes["hidden"]; !hidden {
		t.Error("the native option prototype must not be an active listbox choice")
	}
	for _, want := range []catalog.Binding{
		{Node: "trigger", Kind: "attribute", Name: "id", Prop: "id"},
		{Node: "label", Kind: "attribute", Name: "for", Prop: "id"},
		{Node: "popup", Kind: "attribute", Name: "id", Prop: "popupId"},
		{Node: "trigger", Kind: "attribute", Name: "aria-controls", Prop: "popupId"},
		{Node: "trigger", Kind: "attribute", Name: "popovertarget", Prop: "popupId"},
		{Node: "control", Kind: "property", Name: "value", Prop: "value"},
		{Node: "control", Kind: "attribute", Name: "name", Prop: "name"},
		{Node: "control", Kind: "property", Name: "disabled", Prop: "disabled"},
		{Node: "trigger", Kind: "property", Name: "disabled", Prop: "disabled"},
		{Node: "control", Kind: "property", Name: "required", Prop: "required"},
		{Node: "trigger", Kind: "attribute", Name: "aria-required", Prop: "required"},
	} {
		found := false
		for _, binding := range r.Bindings {
			found = found || binding == want
		}
		if !found {
			t.Errorf("missing semantic relationship: %+v", want)
		}
	}
	for _, slot := range r.Slots {
		if slot.Name == "label" && slot.Required {
			t.Error("external field labels must allow omitting the owned label content")
		}
		if slot.Name == "options" && !slot.Required {
			t.Error("native options are required")
		}
	}
	for _, prop := range r.Props {
		if (prop.Name == "id" || prop.Name == "popupId") && (!prop.Required || prop.Type != "string") {
			t.Errorf("%s must be an explicit required unique ID", prop.Name)
		}
		if prop.Name == "value" && (prop.Required || len(prop.Default) > 0) {
			t.Error("omitted value must preserve option-owned initial selection")
		}
	}
	for _, event := range r.Events {
		if (event.Name == "input" || event.Name == "change") && (event.Target != "control" || event.State["value"] != "value") {
			t.Error("selection notifications must read the native select value")
		}
		if event.Name == "toggle" && (event.Target != "popup" || len(event.State) != 0) {
			t.Error("popup toggle is a notification, not fabricated model state")
		}
	}
}

func TestSelectListStatesFollowTheVisibleTriggerAndNativeProxy(t *testing.T) {
	r, ok := catalog.Find("select-list")
	if !ok {
		t.Fatal("missing select-list")
	}
	for _, state := range []string{"disabled", "focusVisible", "hover", "expanded"} {
		if r.UI.Parts["trigger"].State[state].Source.Node != "trigger" {
			t.Errorf("trigger %s state must reflect the visible trigger", state)
		}
	}
	expanded := r.UI.Parts["trigger"].State["expanded"].Source
	if expanded.Attribute != "aria-expanded" || expanded.Value == nil || *expanded.Value != "true" {
		t.Error("expanded state must read aria-expanded=true")
	}
	for _, state := range []string{"disabled", "invalid", "required"} {
		source := r.UI.Parts["control"].State[state].Source
		if source.Node != "control" || source.Pseudo != state {
			t.Errorf("proxy %s must retain native state", state)
		}
	}
	open := r.UI.Parts["popup"].State["open"].Source
	if open.Node != "popup" || open.Pseudo != "popover-open" {
		t.Error("popup state must follow native popover visibility")
	}
	for state, attribute := range map[string]string{"selected": "aria-selected", "disabled": "aria-disabled", "highlighted": "data-highlighted"} {
		source := r.UI.Parts["option"].State[state].Source
		if source.Node != "option" || source.Attribute != attribute || source.Value == nil || *source.Value != "true" {
			t.Errorf("projected option %s must compare %s=true", state, attribute)
		}
	}
}
