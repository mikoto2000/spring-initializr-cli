package main

import "testing"

func TestSelectedDisplayLines(t *testing.T) {
    selected := map[string]bool{"web": true, "data-jpa": true, "security": false}
    catalog := map[string]depOption{
        "web":      {ID: "web", Name: "Spring Web", Group: "Web"},
        "data-jpa": {ID: "data-jpa", Name: "Spring Data JPA", Group: "SQL"},
    }
    lines := selectedDisplayLines(selected, catalog)
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d", len(lines))
    }
    // They are sorted by ID via selectedIDs, so data-jpa then web
    if lines[0] != "Spring Data JPA (data-jpa) [SQL]" {
        t.Fatalf("bad line[0]: %q", lines[0])
    }
    if lines[1] != "Spring Web (web) [Web]" {
        t.Fatalf("bad line[1]: %q", lines[1])
    }
}

func TestJoinSelected(t *testing.T) {
    selected := map[string]bool{"b": true, "a": true, "c": false}
    s := joinSelected(selected)
    if s != "a,b" {
        t.Fatalf("joinSelected = %q; want a,b", s)
    }
}

