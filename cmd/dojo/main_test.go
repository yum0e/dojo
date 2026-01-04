package main

import (
	"strings"
	"testing"
)

func TestView(t *testing.T) {
	m := model{}
	view := m.View()

	// Print the actual output so we can see it
	t.Logf("View output:\n%s", view)

	// Basic assertions
	if !strings.Contains(view, "DOJO") {
		t.Error("expected view to contain 'DOJO'")
	}
	if !strings.Contains(view, "jj workspaces") {
		t.Error("expected view to contain 'jj workspaces'")
	}
	if !strings.Contains(view, "q") {
		t.Error("expected view to contain quit instruction")
	}
}
