# jorge
A personal (small + opinionated) site generator with [org-mode](https://orgmode.org/) (and markdown) support.

## Installation
Download the [latest release binary](https://github.com/facundoolano/jorge/releases/latest) for your platform, for example:

    $ wget https://github.com/facundoolano/jorge/releases/latest/download/jorge-darwin-amd64  \
        -O jorge && chmod +x jorge && sudo mv jorge /usr/local/bin

Alternatively, install with go (make sure that `$GOPATH/bin` is in your path):

    $ go install github.com/facundoolano/jorge@latest


ArchLinux users can use an [AUR helper](https://wiki.archlinux.org/title/AUR_helpers), such as `yay`, to install `jorge` directly from [AUR](https://aur.archlinux.org/packages/jorge-git):

    $ yay -S jorge

## Example usage

Create a new website with `jorge init`:

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
```

This initializes a new project with default configuration, styles and layouts, and a couple of sample posts.
(You can, of course, use a different site structure or just skip the init command altogether).

To preview your site locally, use `jorge serve`:

```bash
$ cd myblog
$ jorge serve
wrote target/feed.xml
wrote target/blog/goodbye-markdown/index.html
wrote target/blog/my-first-post/index.html
wrote target/blog/hello-org/index.html
wrote target/blog/index.html
wrote target/index.html
wrote target/blog/tags/index.html
serving at http://localhost:4001
```

The site is renders the files found at `src/` in the `target/` directory.
You can add new pages by just adding files to `src/` but, for the common case of adding blog posts,
the `jorge post` creates files with the proper defaults:

```
$ jorge post "My First Post"
added src/blog/my-first-post.org
$ cat src/blog/my-first-post.org
---
title: My First Post
date: 2024-02-21 13:39:59
layout: post
lang: en
tags: []
draft: true
---
#+OPTIONS: toc:nil num:nil
#+LANGUAGE: en
```

(Posts are created as .org files by default, but you can chage it to prefer markdown or another text format).

If you still have `jorge serve` running, you can see the new post by browsing to `http://localhost:4001/blog/my-first-post`. You can then add some content and the browser tab will automatically refresh to reflect your changes:

```bash
$ cat >> src/blog/my-first-post.org <<EOF
*** Hello world!

this is my *first* post.
EOF
```

Posts created with `jorge post` are drafts by default. Remove the `draft: true` to mark it ready for publication:

``` bash
$ sed -i '' '/^draft: true$/d' src/blog/my-first-post.org
```

Finally, you can render a minified version of your site with `jorge build`:

```
$ jorge build
  wrote target/index.html
  wrote target/assets/css/main.css
  wrote target/blog/hello/index.html
  wrote target/blog/my-first-post/index.html
  wrote target/feed.xml
  wrote target/tags/index.html
```

And that's about it. For more details see the:

  - [Tutorial](https://jorge.olano.dev#tutorial)
  - [Development blog](https://jorge.olano.dev#devlog)

## Built with jorge

* [jorge docs](https://jorge.olano.dev/)
* [olano.dev](https://olano.dev/)
* [Schizophrenic.io](https://schizophrenic.io/)


## Acknowledgements

jorge started as a Go learning project and was largely inspired by [Jekyll](https://jekyllrb.com/). Most of the heavy lifting is done by external libraries:

* [osteele/liquid](https://github.com/osteele/liquid) to render liquid templates. Some Jekyll-specific filters were copied from [osteele/gojekyll](https://github.com/osteele/gojekyll/).
* [niklasfasching/go-org](https://github.com/niklasfasching/go-org) to render org-mode files as HTML.
* [yuin/goldmark](https://github.com/yuin/goldmark) to render Markdown as HTML.
* [go-yaml](https://github.com/go-yaml/yaml) to parse YAML files and template headers.
* [tdewolff/minify](https://github.com/tdewolff/minify) to minify HTML, CSS, XML and JavaScript files.
