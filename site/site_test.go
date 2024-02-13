package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndRenderTemplates(t *testing.T) {
	root, src, layouts := newProject()
	defer os.RemoveAll(root)

	// add two layouts
	content := `---
---
<html>
<head><title>{{page.title}}</title></head>
<body>
{{content}}
</body>
</html>`
	file := newFile(layouts, "base.html", content)
	defer os.Remove(file.Name())

	content = `---
layout: base
---
<h1>{{page.title}}</h1>
<h2>{{page.subtitle}}</h2>
{{content}}`
	file = newFile(layouts, "post.html", content)
	defer os.Remove(file.Name())

	// add two posts
	content = `---
layout: post
title: hello world!
subtitle: my first post
date: 2024-01-01
---
<p>Hello world!</p>`
	file = newFile(src, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
layout: post
title: goodbye!
subtitle: my last post
date: 2024-02-01
---
<p>goodbye world!</p>`
	file = newFile(src, "goodbye.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
layout: base
title: about
---
<p>about this site</p>`
	file = newFile(src, "about.html", content)
	defer os.Remove(file.Name())

	// add a static file (no front matter)
	content = `go away!`
	file = newFile(src, "robots.txt", content)

	site, err := Load(src, layouts)

	assertEqual(t, err, nil)

	assertEqual(t, len(site.posts), 2)
	assertEqual(t, len(site.pages), 1)
	assertEqual(t, len(site.layouts), 2)

	_, ok := site.layouts["base"]
	assert(t, ok)
	_, ok = site.layouts["post"]
	assert(t, ok)

	hello := site.posts[1]
	content, err = site.Render(&hello)
	assertEqual(t, err, nil)
	assertEqual(t, content, `<html>
<head><title>hello world!</title></head>
<body>
<h1>hello world!</h1>
<h2>my first post</h2>
<p>Hello world!</p>
</body>
</html>`)

	goodbye := site.posts[0]
	content, err = site.Render(&goodbye)
	assertEqual(t, err, nil)
	assertEqual(t, content, `<html>
<head><title>goodbye!</title></head>
<body>
<h1>goodbye!</h1>
<h2>my last post</h2>
<p>goodbye world!</p>
</body>
</html>`)

	about := site.pages[0]
	content, err = site.Render(&about)
	assertEqual(t, err, nil)
	assertEqual(t, content, `<html>
<head><title>about</title></head>
<body>
<p>about this site</p>
</body>
</html>`)

}

func TestRenderArchive(t *testing.T) {
	// TODO
}

func TestRenderTags(t *testing.T) {
	// TODO
}

func TestRenderDataFile(t *testing.T) {
	// TODO
}

// ------ HELPERS --------

func newProject() (string, string, string) {
	projectDir, _ := os.MkdirTemp("", "root")
	layoutsDir := filepath.Join(projectDir, "layouts")
	srcDir := filepath.Join(projectDir, "src")
	os.Mkdir(layoutsDir, 0777)
	os.Mkdir(filepath.Join(projectDir, "src"), 0777)

	return projectDir, layoutsDir, srcDir
}

func newFile(dir string, filename string, contents string) *os.File {
	path := filepath.Join(dir, filename)
	file, _ := os.Create(path)
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
