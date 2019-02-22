package main

import (
	"testing"
	"github.com/jmyounker/mustache"
)

func TestRender(t *testing.T) {
	tmpl, err := mustache.ParseString("{{m}}{{p}}")
	if err != nil {
		t.Fail()
	}
	m := map[string]string{"m":"mv"}
	e := tmpl.Render(false, m)
	if e != "mv" {
		t.Fatalf("%s", e)
	}
}