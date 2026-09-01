package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestProbe(t *testing.T) {
	m := model{
		state: stateActions,
		width: 80,
		profiles: []Profile{
			{Name: "alpha", User: "root", Host: "a.example.com", Port: 22},
			{Name: "beta", User: "root", Host: "b.example.com", Port: 2222},
			{Name: "gamma", User: "root", Host: "c.example.com", Port: 22},
		},
		cursor: 1,
	}
	for i, line := range strings.Split(m.listView(), "\n") {
		fmt.Printf("%2d| %s\n", i, line)
	}
}
