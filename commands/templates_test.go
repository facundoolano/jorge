package commands

import (
	"strings"
	"testing"
)

func TestExtractFrontMatter(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>`

	outContent, yaml, err := extractFrontMatter(strings.NewReader(input))
	assertEqual(t, err, nil)
	assertEqual(t, string(outContent), "<p>Hello World!</p>")
	assertEqual(t, yaml["title"], "my new post")
	assertEqual(t, yaml["subtitle"], "a blog post")
	assertEqual(t, yaml["tags"].([]interface{})[0], "software")
	assertEqual(t, yaml["tags"].([]interface{})[1], "web")
}

func TestNonFrontMatterDelimiter(t *testing.T) {
	// not identified as front matter, leaving file as is
	input := `+++
title: my new post
subtitle: a blog post
+++
<p>Hello World!</p>`

	out, yaml, err := extractFrontMatter(strings.NewReader(input))

	assertEqual(t, string(out), input)
	assertEqual(t, err, nil)
	assertEqual(t, len(yaml), 0)

	// not first thing in file, leaving as is
	input = `#+OPTIONS: toc:nil num:nil
---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
---
<p>Hello World!</p>`

	out, yaml, err = extractFrontMatter(strings.NewReader(input))

	assertEqual(t, string(out), input)
	assertEqual(t, err, nil)
	assertEqual(t, len(yaml), 0)
}

func TestInvalidFrontMatterYaml(t *testing.T) {
	input := `---
title: my new post
subtitle: a blog post
tags: ["software", "web"]
`

	_, _, err := extractFrontMatter(strings.NewReader(input))
	assertEqual(t, err.Error(), "front matter not closed")

	input = `---
title
tags: ["software", "web"]
---
<p>Hello World!</p>`

	_, _, err = extractFrontMatter(strings.NewReader(input))
	msg := strings.Split(err.Error(), ":")[0]
	assertEqual(t, msg, "invalid yaml")
}

func TestRenderHtml(t *testing.T) {
	// TODO
}

func TestRenderOrg(t *testing.T) {
	// TODO
}

// TODO move to assert package
func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}
