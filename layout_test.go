package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func layoutModel(st state, cursor int) model {
	m := model{
		state:  st,
		width:  70,
		cursor: cursor,
		profiles: []Profile{
			{Name: "db", Group: "production", User: "root", Host: "db.example.com", Port: 22},
			{Name: "web", Group: "production", User: "root", Host: "web.example.com", Port: 2222},
			{Name: "web", Group: "staging", User: "deploy", Host: "stg.example.com", Port: 22},
			{Name: "laptop", User: "me", Host: "192.168.1.10", Port: 22},
		},
	}
	sortProfiles(m.profiles)
	return m
}

// profileRows drives mouse hit testing but duplicates the geometry of listView,
// so the two have to agree, collection headers included.
func TestProfileRowsMatchListView(t *testing.T) {
	for _, st := range []state{stateList, stateActions} {
		m := layoutModel(st, 2)
		lines := strings.Split(m.listView(), "\n")
		for i, top := range m.profileRows() {
			name := m.profiles[i].Name
			if top+1 >= len(lines) || !strings.Contains(lines[top+1], name) {
				t.Errorf("state %d: profile %d (%s) not drawn at row %d", st, i, name, top+1)
			}
		}
	}
}

func TestClickSelectsProfile(t *testing.T) {
	m := layoutModel(stateList, 0)
	click := func(y int) model {
		got, _ := m.handleClick(tea.MouseMsg{Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
		return got.(model)
	}

	for i, top := range m.profileRows() {
		for _, y := range []int{top, top + cardHeight - 1} {
			if c := click(y).cursor; c != i {
				t.Errorf("click at row %d: cursor = %d, want %d", y, c, i)
			}
		}
	}

	if got := click(listTop - 1); got.cursor != 0 || got.state != stateList {
		t.Error("click on a collection header changed the selection")
	}
}

func TestSortGroupsCollectionsFirst(t *testing.T) {
	m := layoutModel(stateList, 0)
	var got []string
	for _, p := range m.profiles {
		got = append(got, p.GroupLabel()+"/"+p.Name)
	}
	want := "production/db production/web staging/web Ungrouped/laptop"
	if strings.Join(got, " ") != want {
		t.Errorf("order = %q, want %q", strings.Join(got, " "), want)
	}
}

func TestNameTakenIsPerCollection(t *testing.T) {
	m := layoutModel(stateList, 0)
	if !nameTaken(m.profiles, Profile{Name: "web", Group: "production"}, -1) {
		t.Error("duplicate name within a collection not detected")
	}
	if nameTaken(m.profiles, Profile{Name: "web", Group: "dev"}, -1) {
		t.Error("same name in a different collection must be allowed")
	}
	if nameTaken(m.profiles, Profile{Name: "web", Group: "production"}, 1) {
		t.Error("the profile being edited must not clash with itself")
	}
}
