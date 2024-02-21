# jorge
A personal (small + opinionated) site generator with [org-mode](https://orgmode.org/) (and markdown) support.

## Installation
Download the [latest release binary](https://github.com/facundoolano/jorge/releases/latest) for your platform, for example:

    $ wget https://github.com/facundoolano/jorge/releases/latest/download/jorge-darwin-amd64  \
        -O jorge && chmod +x jorge && mv jorge /usr/local/bin

Alternatively, install with go:

    $ go install github.com/facundoolano/jorge@latest

## Usage

```bash
$ jorge init myblog
site name: My Blog
site url: https://myblog.olano.dev
author: Facundo Olano

added myblog/.gitignore
added myblog/includes/post_preview.html
added myblog/layouts/base.html
added myblog/layouts/default.html
added myblog/layouts/post.html
added myblog/src/assets/css/main.css
added myblog/src/blog/goodbye-markdown.md
added myblog/src/blog/hello-org.org
added myblog/src/blog/index.html
added myblog/src/blog/tags.html
added myblog/src/feed.xml
added myblog/src/index.html

$ cd myblog
$ jorge post "My First Post"
added src/blog/my-first-post.org

# serve the site locally with live reload
$ jorge serve
wrote target/feed.xml
wrote target/blog/goodbye-markdown.html
wrote target/blog/my-first-post.html
wrote target/blog/hello-org.html
wrote target/blog/index.html
wrote target/index.html
wrote target/blog/tags.html
server listening at http://localhost:4001

# browse to the new post
$ open http://localhost:4001/blog/my-first-post

# add some content
$ cat >> src/blog/my-first-post.org <<EOF
*** Hello world!

this is my *first* post.
EOF

$ jorge build
  wrote target/index.html
  wrote target/assets/css/main.css
  wrote target/blog/hello.html
  wrote target/blog/my-first-post.html
  wrote target/feed.xml
  wrote target/tags.html
```

For more details see the:

  - [Tutorial](https://jorge.olano.dev#tutorial)
  - [Docs](https://jorge.olano.dev#docs)
  - [Development blog](https://jorge.olano.dev#blog)

## Acknowledgements

jorge started as a Go learning project and was largely inspired by [Jekyll](https://jekyllrb.com/). Most of the heavy lifting is done by external libraries:

* [osteele/liquid](https://github.com/osteele/liquid) to render liquid templates. Some Jekyll-specific filters were also copied from [osteele/gojekyll](https://github.com/osteele/gojekyll/).
* [niklasfasching/go-org](https://github.com/niklasfasching/go-org) to render org-mode files as HTML.
* [yuin/goldmark](https://github.com/yuin/goldmark) to render Markdown as HTML.
* [go-yaml](https://github.com/go-yaml/yaml) to parse YAML files and template headers.
* [tdewolff/minify](https://github.com/tdewolff/minify) to minify HTML, CSS, XML and JavaScript files.
