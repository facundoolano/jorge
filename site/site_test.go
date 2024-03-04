package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/facundoolano/jorge/config"
)

func TestLoadAndRenderTemplates(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	// add two layouts
	content := `---
---
<html>
<head><title>{{page.title}}</title></head>
<body>
{{content}}
</body>
</html>`
	file := newFile(config.LayoutsDir, "base.html", content)
	defer os.Remove(file.Name())

	content = `---
layout: base
---
<h1>{{page.title}}</h1>
<h2>{{page.subtitle}}</h2>
{{content}}`
	file = newFile(config.LayoutsDir, "post.html", content)
	defer os.Remove(file.Name())

	// add two posts
	content = `---
layout: post
title: hello world!
subtitle: my first post
date: 2024-01-01
---
<p>Hello world!</p>`
	file = newFile(config.SrcDir, "hello.html", content)
	helloPath := file.Name()
	defer os.Remove(helloPath)

	content = `---
layout: post
title: goodbye!
subtitle: my last post
date: 2024-02-01
---
<p>goodbye world!</p>`
	file = newFile(config.SrcDir, "goodbye.html", content)
	goodbyePath := file.Name()
	defer os.Remove(goodbyePath)

	// add a page (no date)
	content = `---
layout: base
title: about
---
<p>about this site</p>`
	file = newFile(config.SrcDir, "about.html", content)
	aboutPath := file.Name()
	defer os.Remove(aboutPath)

	// add a static file (no front matter)
	content = `go away!`
	newFile(config.SrcDir, "robots.txt", content)

	site, err := load(*config)

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

func TestPreviousNext(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	// prev next distinguish between series in different dirs
	// add two different blog dirs and two different page dirs (tutorials)
	blog1 := filepath.Join(config.SrcDir, "blog1")
	blog2 := filepath.Join(config.SrcDir, "blog2")
	tutorial1 := filepath.Join(config.SrcDir, "tutorial1")
	tutorial2 := filepath.Join(config.SrcDir, "tutorial2")
	os.Mkdir(blog1, DIR_RWE_MODE)
	os.Mkdir(blog2, DIR_RWE_MODE)
	os.Mkdir(tutorial1, DIR_RWE_MODE)
	os.Mkdir(tutorial2, DIR_RWE_MODE)

	newFile(blog1, "p1-1.html", `---
date: 2024-01-01
---`)
	newFile(blog1, "p1-2.html", `---
date: 2024-01-02
---`)
	newFile(blog1, "p1-3.html", `---
date: 2024-01-03
---`)

	newFile(blog2, "p2-1.html", `---
date: 2024-02-01
---`)
	newFile(blog2, "p2-2.html", `---
date: 2024-02-02
---`)
	newFile(blog2, "p2-3.html", `---
date: 2024-02-03
---`)

	newFile(tutorial1, "1-first-part.html", `---
---`)
	newFile(tutorial1, "2-another-entry.html", `---
---`)
	newFile(tutorial1, "3-goodbye.html", `---
---`)

	newFile(tutorial2, "index.html", `---
---`)
	newFile(tutorial2, "the-end.html", `---
---`)
	newFile(tutorial2, "another-entry.html", `---
---`)

	site, err := load(*config)
	// helper method to map a filename to its prev next keys (if any)
	getPrevNext := func(dir string, filename string) (interface{}, interface{}) {
		path := filepath.Join(dir, filename)
		templ := site.templates[path]
		return templ.Metadata["previous"], templ.Metadata["next"]
	}
	// tests for posts (sorted reverse chronologically, most recent one first)
	assertEqual(t, err, nil)
	prev, next := getPrevNext(blog1, "p1-3.html")
	assertEqual(t, prev, nil)
	assertEqual(t, next.(map[string]interface{})["url"], "/blog1/p1-2")

	prev, next = getPrevNext(blog1, "p1-2.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/blog1/p1-3")
	assertEqual(t, next.(map[string]interface{})["url"], "/blog1/p1-1")

	prev, next = getPrevNext(blog1, "p1-1.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/blog1/p1-2")
	assertEqual(t, next, nil)

	assertEqual(t, err, nil)
	prev, next = getPrevNext(blog2, "p2-3.html")
	assertEqual(t, prev, nil)
	assertEqual(t, next.(map[string]interface{})["url"], "/blog2/p2-2")

	prev, next = getPrevNext(blog2, "p2-2.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/blog2/p2-3")
	assertEqual(t, next.(map[string]interface{})["url"], "/blog2/p2-1")

	prev, next = getPrevNext(blog2, "p2-1.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/blog2/p2-2")
	assertEqual(t, next, nil)

	// test for pages based on filename
	prev, next = getPrevNext(tutorial1, "1-first-part.html")
	assertEqual(t, prev, nil)
	assertEqual(t, next.(map[string]interface{})["url"], "/tutorial1/2-another-entry")

	prev, next = getPrevNext(tutorial1, "2-another-entry.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/tutorial1/1-first-part")
	assertEqual(t, next.(map[string]interface{})["url"], "/tutorial1/3-goodbye")

	prev, next = getPrevNext(tutorial1, "3-goodbye.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/tutorial1/2-another-entry")
	assertEqual(t, next, nil)

	// ensure alphabetical and index skipped
	prev, next = getPrevNext(tutorial2, "another-entry.html")
	assertEqual(t, prev, nil)
	assertEqual(t, next.(map[string]interface{})["url"], "/tutorial2/the-end")

	prev, next = getPrevNext(tutorial2, "the-end.html")
	assertEqual(t, prev.(map[string]interface{})["url"], "/tutorial2/another-entry")
	assertEqual(t, next, nil)
}

func TestRenderArchive(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	content := `---
title: hello world!
date: 2024-01-01
---
<p>Hello world!</p>`
	file := newFile(config.SrcDir, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: goodbye!
date: 2024-02-01
---
<p>goodbye world!</p>`
	file = newFile(config.SrcDir, "goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: an oldie!
date: 2023-01-01
---
<p>oldie</p>`
	file = newFile(config.SrcDir, "an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
<ul>{% for post in site.posts %}
<li>{{ post.date | date: "%Y-%m-%d" }} <a href="{{ post.url }}">{{post.title}}</a></li>{%endfor%}
</ul>`

	file = newFile(config.SrcDir, "about.html", content)
	defer os.Remove(file.Name())

	site, _ := load(*config)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<ul>
<li>2024-02-01 <a href="/goodbye">goodbye!</a></li>
<li>2024-01-01 <a href="/hello">hello world!</a></li>
<li>2023-01-01 <a href="/an-oldie">an oldie!</a></li>
</ul>`)
}

func TestRenderTags(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	content := `---
title: hello world!
date: 2024-01-01
tags: [web, software]
---
<p>Hello world!</p>`
	file := newFile(config.SrcDir, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: goodbye!
date: 2024-02-01
tags: [web]
---
<p>goodbye world!</p>`
	file = newFile(config.SrcDir, "goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: an oldie!
date: 2023-01-01
tags: [software]
---
<p>oldie</p>`
	file = newFile(config.SrcDir, "an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
{% for tag in site.tags | keys | sort %}<h1>{{tag}}</h1>{% for post in site.tags[tag] %}
{{post.title}}
{% endfor %}
{% endfor %}
`

	file = newFile(config.SrcDir, "about.html", content)
	defer os.Remove(file.Name())

	site, _ := load(*config)
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
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	content := `---
title: "1. hello world!"
---
<p>Hello world!</p>`
	file := newFile(config.SrcDir, "01-hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: "3. goodbye!"
---
<p>goodbye world!</p>`
	file = newFile(config.SrcDir, "03-goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: "2. an oldie!"
---
<p>oldie</p>`
	file = newFile(config.SrcDir, "02-an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
<ul>{% for page in site.pages %}
<li><a href="{{ page.url }}">{{page.title}}</a></li>{%endfor%}
</ul>`

	file = newFile(config.SrcDir, "index.html", content)
	defer os.Remove(file.Name())

	site, _ := load(*config)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<ul>
<li><a href="/01-hello">1. hello world!</a></li>
<li><a href="/02-an-oldie">2. an oldie!</a></li>
<li><a href="/03-goodbye">3. goodbye!</a></li>
</ul>`)
}

func TestRenderArchiveWithExcerpts(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	content := `---
title: hello world!
date: 2024-01-01
tags: [web, software]
---
<p>the intro paragraph</p>
<p> and another paragraph>`
	file := newFile(config.SrcDir, "hello.html", content)
	defer os.Remove(file.Name())

	content = `---
title: goodbye!
date: 2024-02-01
tags: [web]
excerpt: an overridden excerpt
---
<p>goodbye world!</p>
<p> and another paragraph</p>`
	file = newFile(config.SrcDir, "goodbye.html", content)
	defer os.Remove(file.Name())

	content = `---
title: an oldie!
date: 2023-01-01
tags: [software]
---
`
	file = newFile(config.SrcDir, "an-oldie.html", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
{% for post in site.posts %}
{{post.title}} - {{post.excerpt}}
{% endfor %}
`

	file = newFile(config.SrcDir, "about.html", content)
	defer os.Remove(file.Name())

	site, _ := load(*config)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, strings.TrimSpace(string(output)), `goodbye! - an overridden excerpt

hello world! - the intro paragraph

an oldie! -`)
}

func TestRenderDataFile(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	content := `
- name: feedi
  url: https://github.com/facundoolano/feedi
- name: jorge
  url: https://github.com/facundoolano/jorge
`
	file := newFile(config.DataDir, "projects.yml", content)
	defer os.Remove(file.Name())

	// add a page (no date)
	content = `---
---
<ul>{% for project in site.data.projects %}
<li><a href="{{ project.url }}">{{project.name}}</a></li>{%endfor%}
</ul>`

	file = newFile(config.SrcDir, "projects.html", content)
	defer os.Remove(file.Name())

	site, _ := load(*config)
	output, err := site.render(site.templates[file.Name()])
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<ul>
<li><a href="https://github.com/facundoolano/feedi">feedi</a></li>
<li><a href="https://github.com/facundoolano/jorge">jorge</a></li>
</ul>`)
}

func TestBuildTarget(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	// add base layout
	content := `---
---
<html>
<head><title>{{page.title}}</title></head>
<body>
{{content}}
</body>
</html>`
	newFile(config.LayoutsDir, "base.html", content)

	// add org post
	content = `---
layout: base
title: p1 - hello world!
date: 2024-01-01
---
* Hello world!`
	newFile(config.SrcDir, "p1.org", content)

	// add markdown post
	content = `---
layout: base
title: p2 - goodbye world!
date: 2024-01-02
---
# Goodbye world!`
	newFile(config.SrcDir, "p2.md", content)

	// add index page
	content = `---
layout: base
---
<ul>{% for post in site.posts %}
<li>{{post.title}}</li>{%endfor%}
</ul>`
	newFile(config.SrcDir, "index.html", content)

	// build site
	site, err := load(*config)
	assertEqual(t, err, nil)
	err = site.build()
	assertEqual(t, err, nil)

	// test target files generated
	_, err = os.Stat(filepath.Join(config.TargetDir, "p1.html"))
	assertEqual(t, err, nil)
	_, err = os.Stat(filepath.Join(config.TargetDir, "p2.html"))
	assertEqual(t, err, nil)

	// test index includes p1 and p2
	output, err := os.ReadFile(filepath.Join(config.TargetDir, "index.html"))
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html><head><title></title></head>
<body>
<ul>
<li>p2 - goodbye world!</li>
<li>p1 - hello world!</li>
</ul>

</body></html>`)
}

func TestBuildWithDrafts(t *testing.T) {
	config := newProject()
	defer os.RemoveAll(config.RootDir)

	// add base layout
	content := `---
---
<html>
<head><title>{{page.title}}</title></head>
<body>
{{content}}
</body>
</html>`
	newFile(config.LayoutsDir, "base.html", content)

	// add org post
	content = `---
layout: base
title: p1 - hello world!
date: 2024-01-01
---
* Hello world!`
	newFile(config.SrcDir, "p1.org", content)

	// add markdown post, make it draft
	content = `---
layout: base
title: p2 - goodbye world!
date: 2024-01-02
draft: true
---
# Goodbye world!`
	newFile(config.SrcDir, "p2.md", content)

	// add index page
	content = `---
layout: base
---
<ul>{% for post in site.posts %}
<li>{{post.title}}</li>{%endfor%}
</ul>`
	newFile(config.SrcDir, "index.html", content)

	// build site with drafts
	config.IncludeDrafts = true
	site, err := load(*config)
	assertEqual(t, err, nil)
	err = site.build()
	assertEqual(t, err, nil)

	// test target files generated
	_, err = os.Stat(filepath.Join(config.TargetDir, "p1.html"))
	assertEqual(t, err, nil)
	_, err = os.Stat(filepath.Join(config.TargetDir, "p2.html"))
	assertEqual(t, err, nil)

	// test index includes p1 and p2
	output, err := os.ReadFile(filepath.Join(config.TargetDir, "index.html"))
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html><head><title></title></head>
<body>
<ul>
<li>p2 - goodbye world!</li>
<li>p1 - hello world!</li>
</ul>

</body></html>`)

	// build site WITHOUT drafts
	config.IncludeDrafts = false
	site, err = load(*config)
	assertEqual(t, err, nil)
	err = site.build()
	assertEqual(t, err, nil)

	// test only non drafts generated
	_, err = os.Stat(filepath.Join(config.TargetDir, "p1.html"))
	assertEqual(t, err, nil)
	_, err = os.Stat(filepath.Join(config.TargetDir, "p2.html"))
	assert(t, os.IsNotExist(err))

	// test index includes p1 but  NOT p2
	output, err = os.ReadFile(filepath.Join(config.TargetDir, "index.html"))
	assertEqual(t, err, nil)
	assertEqual(t, string(output), `<html><head><title></title></head>
<body>
<ul>
<li>p1 - hello world!</li>
</ul>

</body></html>`)
}

// ------ HELPERS --------

func newProject() *config.Config {
	projectDir, _ := os.MkdirTemp("", "root")
	layoutsDir := filepath.Join(projectDir, "layouts")
	srcDir := filepath.Join(projectDir, "src")
	dataDir := filepath.Join(projectDir, "data")
	os.Mkdir(layoutsDir, DIR_RWE_MODE)
	os.Mkdir(srcDir, DIR_RWE_MODE)
	os.Mkdir(dataDir, DIR_RWE_MODE)

	config, _ := config.Load(projectDir)
	config.Minify = false

	return config
}

func newFile(dir string, filename string, contents string) *os.File {
	path := filepath.Join(dir, filename)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
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
