package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jmyounker/jtools/internal/mustache"
)

var inputTests = []struct{ in, tmpl, want string }{
	{"", "foo", ""},
	{"{}", "", "\n"},
	{"{\"a\": \"foo\"}", "{{a}}", "foo\n"},
	{"{\"a\": \"foo\"}{\"a\":\"bar\"}", "{{a}}", "foo\nbar\n"},
	{"[\"foo\"]", "{{1}}", "foo\n"},
	{"\"foo\"", "{{.}}", "foo\n"},
	{"true", "{{.}}", "true\n"},
	{"42", "{{.}}", "42\n"},
	{"null", "{{.}}", "\n"},
	{"null", "{{{.}}}", "\n"},
	{"{\"a\":[1,2]}", "{{{a}}}", "[1,2]\n"},
	{"{\"a\":[\"b\",\"c\"]}", "{{{a}}}", "[\"b\",\"c\"]\n"},
	{"{\"a\":[\"b\",\"c\"]}", "{{{a.1}}}", "b\n"},
	{"{\"a\":[[\"b1\",\"b2\"],[\"c1\",\"c2\"]]}", "{{#a}}{{2}}{{/a}}", "b2c2\n"},
	{"{\"a\":{\"b\":\"c\"}}", "{{{a}}}", "{\"b\":\"c\"}\n"},
}

func TestExpandInput(t *testing.T) {
	for _, tc := range inputTests {
		in := strings.NewReader(tc.in)
		tmpl, err := mustache.ParseString(tc.tmpl)
		assertNoError(t, err)
		out := bytes.Buffer{}
		outFact := &staticWriterFactory{&out}
		assertNoError(t, expand(in, outFact, &staticTemplateFactory{tmpl}, true, false))
		if out.String() != tc.want {
			t.Fatalf("Expected %q but got %q", tc.want, out.String())
		}
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error but got: %s", err)
	}
}
