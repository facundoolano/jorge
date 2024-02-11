package templates

import (
	"os"
	"strings"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>
`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)

	assertEqual(t, templ.Ext(), ".html")
	assertEqual(t, templ.Metadata["title"], "my new post")
	assertEqual(t, templ.Metadata["subtitle"], "a blog post")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[0], "software")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[1], "web")

	content, err := templ.Render()
	assertEqual(t, err, nil)
	assertEqual(t, string(content), "<p>Hello World!</p>\n")
}

func TestNonTemplate(t *testing.T) {
	// not identified as front matter, leaving file as is
	input := `+++
title: my new post
subtitle: a blog post
+++
<p>Hello World!</p>`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(file.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ, (*Template)(nil))

	// not first thing in file, leaving as is
	input = `#+OPTIONS: toc:nil num:nil
---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>`

	file = newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err = Parse(file.Name())
	assertEqual(t, err, nil)
	assertEqual(t, templ, (*Template)(nil))
}

func TestInvalidFrontMatter(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
`
	file := newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err := Parse(file.Name())

	assertEqual(t, err.Error(), "front matter not closed")

	input = `---
title
tags: ["software", "web"]
---
<p>Hello World!</p>`

	file = newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err = Parse(file.Name())
	assert(t, strings.Contains(err.Error(), "invalid yaml"))
}

func TestRenderLiquid(t *testing.T) {
	// TODO
}

func TestRenderOrg(t *testing.T) {
	// TODO
}

// ------ HELPERS --------

func newFile(name string, contents string) *os.File {
	file, _ := os.CreateTemp("", name)
	file.WriteString(contents)
	return file
}

// TODO move to assert package
func assert(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Fatalf("%v is false", cond)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}
