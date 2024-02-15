package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndRenderTemplates(t *testing.T) {
	root, layouts, src := newProject()
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
	helloPath := file.Name()
	defer os.Remove(helloPath)

	content = `---
layout: post
title: goodbye!
subtitle: my last post
date: 2024-02-01
---
<p>goodbye world!</p>`
	file = newFile(src, "goodbye.html", content)
	goodbyePath := file.Name()
	defer os.Remove(goodbyePath)

	// add a page (no date)
	content = `---
layout: base
title: about
---
<p>about this site</p>`
	file = newFile(src, "about.html", content)
	aboutPath := file.Name()
	defer os.Remove(aboutPath)

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

	output, err := site.render(site.templates[helloPath])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html>
<head><title>hello world!</title></head>
<body>
<h1>hello world!</h1>
<h2>my first post</h2>
<p>Hello world!</p>
</body>
</html>`)

	output, err = site.render(site.templates[goodbyePath])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html>
<head><title>goodbye!</title></head>
<body>
<h1>goodbye!</h1>
<h2>my last post</h2>
<p>goodbye world!</p>
</body>
</html>`)

	output, err = site.render(site.templates[aboutPath])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html>
<head><title>about</title></head>
<body>
<p>about this site</p>
</body>
</html>`)

}

func TestRenderArchive(t *testing.T) {
	root, layouts, src := newProject()
	defer os.RemoveAll(root)

	content := `---
title: hello world!
date: 2024-01-01
---
<p>Hello world!</p>`
	file := newFile(src, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: goodbye!
date: 2024-02-01
---
<p>goodbye world!</p>`
	file = newFile(src, "goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: an oldie!
date: 2023-01-01
---
<p>oldie</p>`
	file = newFile(src, "an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
<ul>{% for post in site.posts %}
<li>{{ post.date | date: "%Y-%m-%d" }} <a href="{{ post.url }}">{{post.title}}</a></li>{%endfor%}
</ul>`

	file = newFile(src, "about.html", content)
	defer os.Remove(file.Name())

	site, err := Load(src, layouts)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<ul>
<li>2024-02-01 <a href="/goodbye">goodbye!</a></li>
<li>2024-01-01 <a href="/hello">hello world!</a></li>
<li>2023-01-01 <a href="/an-oldie">an oldie!</a></li>
</ul>`)
}

func TestRenderTags(t *testing.T) {
	root, layouts, src := newProject()
	defer os.RemoveAll(root)

	content := `---
title: hello world!
date: 2024-01-01
tags: [web, software]
---
<p>Hello world!</p>`
	file := newFile(src, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: goodbye!
date: 2024-02-01
tags: [web]
---
<p>goodbye world!</p>`
	file = newFile(src, "goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: an oldie!
date: 2023-01-01
tags: [software]
---
<p>oldie</p>`
	file = newFile(src, "an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
{% for tag in site.tags %}<h1>{{tag[0]}}</h1>{% for post in tag[1] %}
{{post.title}}
{% endfor %}
{% endfor %}
`

	file = newFile(src, "about.html", content)
	defer os.Remove(file.Name())

	site, err := Load(src, layouts)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<h1>software</h1>
hello world!

an oldie!

<h1>web</h1>
goodbye!

hello world!

`)
}

func TestRenderPagesInDir(t *testing.T) {
	root, layouts, src := newProject()
	defer os.RemoveAll(root)

	content := `---
title: "1. hello world!"
---
<p>Hello world!</p>`
	file := newFile(src, "01-hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: "3. goodbye!"
---
<p>goodbye world!</p>`
	file = newFile(src, "03-goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: "2. an oldie!"
---
<p>oldie</p>`
	file = newFile(src, "02-an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
<ul>{% for page in site.pages %}
<li><a href="{{ page.url }}">{{page.title}}</a></li>{%endfor%}
</ul>`

	file = newFile(src, "index.html", content)
	defer os.Remove(file.Name())

	site, err := Load(src, layouts)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<ul>
<li><a href="/01-hello">1. hello world!</a></li>
<li><a href="/02-an-oldie">2. an oldie!</a></li>
<li><a href="/03-goodbye">3. goodbye!</a></li>
</ul>`)
}

func TestRenderArchiveWithExcerpts(t *testing.T) {
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
