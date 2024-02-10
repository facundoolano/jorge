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

func TestInvalidFrontMatterDelimiter(t *testing.T) {
	// TODO
}

func TestInvalidFrontMatterYaml(t *testing.T) {
	// TODO
}

func TestFrontMatterNotAtTop(t *testing.T) {
	// TODO
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
		t.Fatalf("%s != %s", a, b)
	}
}
