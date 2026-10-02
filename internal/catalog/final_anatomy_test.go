package catalog_test

import (
	"html-ui/internal/catalog"
	"testing"
)

func TestFinalParityAnatomy(t *testing.T) {
	for name, parts := range map[string][]string{
		"otp-field":     {"control", "cells", "cell", "character", "caret"},
		"split-view":    {"root", "start", "separator", "grip", "end"},
		"toast-message": {"root", "announcer", "surface", "title", "content", "actions", "close"},
	} {
		c, ok := catalog.Find(name)
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		for _, part := range parts {
			if c.UI.Parts[part].Node == "" {
				t.Errorf("%s missing %s", name, part)
			}
		}

		if name == "toast-message" {
			for _, node := range c.Nodes {
				if node.ID == "surface" {
					if _, hidden := node.Attributes["hidden"]; !hidden {
						t.Error("toast surface must start hidden before its adapter mounts")
					}
				}
			}
		}
	}
}

func TestFinalParityStateContracts(t *testing.T) {
	for name, parts := range map[string]map[string][]string{
		"otp-field":     {"control": {"disabled", "invalid", "focusVisible"}, "cell": {"disabled", "invalid", "active", "selected"}},
		"split-view":    {"separator": {"disabled", "focusVisible"}},
		"toast-message": {"surface": {"open"}},
	} {
		c, ok := catalog.Find(name)
		if !ok {
			t.Fatal("missing primitive")
		}
		for part, states := range parts {
			for _, state := range states {
				if _, ok := c.UI.Parts[part].State[state]; !ok {
					t.Errorf("%s.%s missing %s", name, part, state)
				}
			}
		}
	}
}
