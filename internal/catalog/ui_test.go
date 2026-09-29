package catalog_test

import (
	"encoding/json"
	"html-ui/internal/catalog"
	"testing"
)

func TestRejectInvalidUIMetadata(t *testing.T) {
	// Each fixture is otherwise a valid one-node native recipe. Invalid UI
	// references must fail rather than letting downstream emitters guess.
	for _, tc := range []struct{ name, ui string }{
		{"missing metadata", `{}`},
		{"unknown node", `{"Parts":{"root":{"node":"absent","styleRole":"action"}}}`},
		{"duplicate node targets", `{"Parts":{"root":{"node":"root","styleRole":"action"},"other":{"node":"root","styleRole":"action"}}}`},
		{"missing root part", `{"Parts":{"other":{"node":"root","styleRole":"action"}}}`},
		{"missing style role", `{"Parts":{"root":{"node":"root"}}}`},
		{"invalid public name", `{"Parts":{"bad name":{"node":"root","styleRole":"action"}}}`},
		{"missing state source", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{}}}}}`},
		{"unknown state source node", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{"source":{"node":"absent","pseudo":"disabled"}}}}}}`},
		{"unknown pseudo", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{"source":{"node":"root","pseudo":"invented"}}}}}}`},
		{"ambiguous source", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{"source":{"node":"root","pseudo":"disabled","attribute":"disabled","present":true}}}}}}`},
		{"missing attribute comparison", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"pressed":{"source":{"node":"root","attribute":"aria-pressed"}}}}}}`},
		{"ambiguous comparison", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"pressed":{"source":{"node":"root","attribute":"aria-pressed","present":true,"value":"true"}}}}}}`},
		{"aria presence is not disabled", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{"source":{"node":"root","attribute":"aria-disabled","present":true}}}}}}`},
		{"inert is not disabled", `{"Parts":{"root":{"node":"root","styleRole":"action","state":{"disabled":{"source":{"node":"root","attribute":"inert","present":true}}}}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var r catalog.Recipe
			data := `{"Name":"example","Status":"native","Nodes":[{"ID":"root","Tag":"button"}],"UI":` + tc.ui + `}`
			if err := json.Unmarshal([]byte(data), &r); err != nil {
				t.Fatal(err)
			}
			if err := catalog.Validate(r); err == nil {
				t.Fatal("invalid UI metadata accepted")
			}
		})
	}
}

func TestRejectNativeDisabledOnSummary(t *testing.T) {
	var r catalog.Recipe
	data := `{"Name":"example","Status":"native","Nodes":[{"ID":"root","Tag":"summary"}],"UI":{"Parts":{"root":{"node":"root","styleRole":"disclosure-trigger","state":{"disabled":{"source":{"node":"root","pseudo":"disabled"}}}}}}}`
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Validate(r); err == nil {
		t.Fatal("summary was given fictional native disabled behavior")
	}
}
