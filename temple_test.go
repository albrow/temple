package temple

import (
	"testing"
)

func TestAddTemplate(t *testing.T) {
	defer reset()
	if err := AddTemplate("test", `Hello, {{ . }}!`); err != nil {
		t.Fatalf("Unexpected error in AddTemplate: %s", err.Error())
	}
	// Get the template from the map
	testTmpl, found := Templates["test"]
	if !found {
		t.Fatal(`Template named "test" was not added to map of Templates`)
	}
	ExpectExecutorOutputs(t, testTmpl, "world", "Hello, world!")
}

func TestAddPartial(t *testing.T) {
	defer reset()
	// Add some Partials and make sure each is added to the map
	partials := map[string]string{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
		// foobarbaz calls on each of the other partials. This tests that
		// partials are associated with all other partials.
		"foobarbaz": `{{ template "partials/foo" }}{{ template "partials/bar" }}{{ template "partials/baz" }}`,
	}
	for name, src := range partials {
		if err := AddPartial(name, src); err != nil {
			t.Fatalf("Unexpected error in AddPartial: %s", err.Error())
		}
		if _, found := Partials[name]; !found {
			t.Errorf(`Partial named "%s" was not added to the map of Partials`, name)
		}
	}
	// The test template calls on each of the four partials. This tests that
	// partials are associated with templates.
	if err := AddTemplate("test", `{{ template "partials/foo" }} {{ template "partials/bar" }} {{ template "partials/baz" }} {{ template "partials/foobarbaz" }}`); err != nil {
		t.Fatalf("Unexpected error in AddTemplate: %s", err.Error())
	}
	testTmpl, found := Templates["test"]
	if !found {
		t.Fatal(`Template named "test" was not added to map of Templates`)
	}
	ExpectExecutorOutputs(t, testTmpl, nil, "foo bar baz foobarbaz")
}

func TestAddLayout(t *testing.T) {
	defer reset()
	// The foo partial will be called on by the header layout, which tests that
	// partials are associated with layouts.
	if err := AddPartial("foo", "foo"); err != nil {
		t.Fatalf("Unexpected error in AddPartial: %s", err.Error())
	}
	if _, found := Partials["foo"]; !found {
		t.Errorf(`Partial named "%s" was not added to the map of Partials`, "foo")
	}
	// The header layout renders a content template (which must be defined by a template using
	// the layout) and calls for the foo partial.
	if err := AddLayout("header", `<h2>{{ template "content" }} {{ template "partials/foo" }}</h2>`); err != nil {
		t.Fatalf("Unexpected error in AddLayout: %s", err.Error())
	}
	if _, found := Layouts["header"]; !found {
		t.Errorf(`Layout named "%s" was not added to the map of Layouts`, "header")
	}
	// The test template defines a content template and attempts to render itself inside the
	// header layout.
	if err := AddTemplate("test", `{{ define "content"}}test{{end}}{{ template "layouts/header" }}`); err != nil {
		t.Fatalf("Unexpected error in AddTemplate: %s", err.Error())
	}
	testTmpl, found := Templates["test"]
	if !found {
		t.Fatal(`Template named "test" was not added to map of Templates`)
	}
	ExpectExecutorOutputs(t, testTmpl, nil, "<h2>test foo</h2>")
}