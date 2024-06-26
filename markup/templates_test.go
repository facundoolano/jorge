package markup

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

	templ, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)

	assertEqual(t, templ.Metadata["title"], "my new post")
	assertEqual(t, templ.Metadata["subtitle"], "a blog post")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[0], "software")
	assertEqual(t, templ.Metadata["tags"].([]interface{})[1], "web")

	content, err := templ.Render()
	assertEqual(t, err, nil)
	assertEqual(t, string(content), "<p>Hello World!</p>")
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

	_, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)

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

	_, err = Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)
}

func TestInvalidFrontMatter(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
`
	file := newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())

	assertEqual(t, err.Error(), "front matter not closed")

	input = `---
title
tags: ["software", "web"]
---
<p>Hello World!</p>`

	file = newFile("test*.html", input)
	defer os.Remove(file.Name())
	_, err = Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assert(t, strings.Contains(err.Error(), "invalid yaml"))
}

func TestRenderLiquid(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<h1>{{ page.title }}</h1>
<h2>{{ page.subtitle }}</h2>
<ul>{% for tag in page.tags %}
<li>{{tag}}</li>{% endfor %}
</ul>`

	file := newFile("test*.html", input)
	defer os.Remove(file.Name())

	templ, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)
	content, err := templ.Render()
	assertEqual(t, err, nil)
	expected := `<h1>my new post</h1>
<h2>a blog post</h2>
<ul>
<li>software</li>
<li>web</li>
</ul>`
	assertEqual(t, string(content), expected)
}

func TestRenderOrg(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
#+OPTIONS: toc:nil num:nil
* My title
** my Subtitle
- list 1
- list 2
`

	file := newFile("test*.org", input)
	defer os.Remove(file.Name())

	templ, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)

	content, err := templ.Render()
	assertEqual(t, err, nil)
	expected := `<h1 id="my-title">
My title
</h1>
<h2 id="my-subtitle">
my Subtitle
</h2>
<ul>
<li>list 1</li>
<li>list 2</li>
</ul>
`
	assertEqual(t, string(content), expected)
}

func TestRenderMarkdown(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
# My title
## my Subtitle
- list 1
- list 2
`

	file := newFile("test*.md", input)
	defer os.Remove(file.Name())

	templ, err := Parse(NewEngine("https://olano.dev", "includes"), file.Name())
	assertEqual(t, err, nil)

	content, err := templ.Render()
	assertEqual(t, err, nil)
	expected := `<h1>My title</h1>
<h2>my Subtitle</h2>
<ul>
<li>list 1</li>
<li>list 2</li>
</ul>
`
	assertEqual(t, string(content), expected)
}

// ------ HELPERS --------

func newFile(path string, contents string) *os.File {
	file, _ := os.CreateTemp("", path)
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
