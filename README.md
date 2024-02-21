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
  added draft src/blog/my-first-post.org

# serve the site locally with live reload
$ jorge serve
  server running at http://localhost:4001/

# browse to the new post
$ open http://localhost:4001/blog/my-first-post

# add some content
$ cat >> src/blog/test.org <<EOF
*** Hello world!

this is my *first* post.
EOF

# remove the draft flag before publishing
$ sed -i '/^draft: true$/d' src/blog/my-first-post.org

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
