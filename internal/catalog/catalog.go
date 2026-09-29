// Package catalog holds the reviewed, embedded primitive definitions.
package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed catalog.json
var data []byte

type Prop struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     json.RawMessage
}
type Scope struct {
	Name string
	Type string
	Prop string
}
type Slot struct {
	Name        string
	Description string
	Required    bool
	Scope       []Scope
}
type Event struct {
	Name        string
	Type        string
	Target      string
	State       map[string]string
	Description string
}
type Node struct {
	ID         string
	Tag        string
	Attributes map[string]string
}
type Binding struct {
	Node string
	Kind string
	Name string
	Prop string
}
type Child struct {
	Parent string
	Node   string
	Slot   string
}
type Feature struct {
	ID     string
	Status string
	Date   string
}
type Recipe struct {
	Name         string
	Title        string
	Summary      string
	Status       string
	RootType     string
	Props        []Prop
	Slots        []Slot
	Events       []Event
	Nodes        []Node
	Bindings     []Binding
	Children     []Child
	Requirements []string
	Styling      []string
	Features     []Feature
	UI           UIContract
}

var recipes = load()

func load() []Recipe {
	var result []Recipe
	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}
	attachUI(result)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// All returns catalog entries in stable name order. Callers must not mutate them.
func All() []Recipe { return recipes }
func Find(name string) (Recipe, bool) {
	i := sort.Search(len(recipes), func(i int) bool { return recipes[i].Name >= name })
	if i < len(recipes) && recipes[i].Name == name {
		return recipes[i], true
	}
	return Recipe{}, false
}

// Validate checks references in a catalog entry; used by catalog maintenance tests.
func Validate(r Recipe) error {
	if r.Name == "" || len(r.Nodes) == 0 || r.Nodes[0].ID != "root" {
		return fmt.Errorf("missing name or root")
	}
	props := map[string]Prop{}
	nodes := map[string]bool{}
	slots := map[string]bool{}
	for _, p := range r.Props {
		if _, ok := props[p.Name]; ok {
			return fmt.Errorf("duplicate prop %s", p.Name)
		}
		props[p.Name] = p
	}
	for _, n := range r.Nodes {
		if nodes[n.ID] {
			return fmt.Errorf("duplicate node %s", n.ID)
		}
		nodes[n.ID] = true
	}
	for _, s := range r.Slots {
		if slots[s.Name] {
			return fmt.Errorf("duplicate slot %s", s.Name)
		}
		slots[s.Name] = true
		for _, v := range s.Scope {
			p, ok := props[v.Prop]
			if !ok || p.Type != v.Type {
				return fmt.Errorf("invalid scope %s", v.Name)
			}
		}
	}
	for _, b := range r.Bindings {
		if _, ok := props[b.Prop]; !ok || !nodes[b.Node] || (b.Kind != "property" && b.Kind != "attribute") {
			return fmt.Errorf("invalid binding %+v", b)
		}
	}
	usedNodes := map[string]bool{}
	usedSlots := map[string]bool{}
	for _, c := range r.Children {
		if !nodes[c.Parent] || (c.Node == "") == (c.Slot == "") {
			return fmt.Errorf("invalid child %+v", c)
		}
		if c.Node != "" {
			if !nodes[c.Node] || usedNodes[c.Node] || c.Node == "root" || c.Node == c.Parent {
				return fmt.Errorf("invalid node insertion %+v", c)
			}
			usedNodes[c.Node] = true
		} else {
			if !slots[c.Slot] || usedSlots[c.Slot] {
				return fmt.Errorf("invalid slot insertion %+v", c)
			}
			usedSlots[c.Slot] = true
		}
	}
	if len(usedNodes) != len(nodes)-1 || len(usedSlots) != len(slots) {
		return fmt.Errorf("unplaced node or slot")
	}
	for _, e := range r.Events {
		if !nodes[e.Target] {
			return fmt.Errorf("invalid event target")
		}
		for p := range e.State {
			if _, ok := props[p]; !ok {
				return fmt.Errorf("invalid event state prop %s", p)
			}
		}
	}
	return validateUI(r)
}
