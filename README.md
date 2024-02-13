# blorg
A presonal (small + opinionated) site generator with org-mode support.

Install from binary:

    $ wget https://github.com/facundoolano/blorg/releases/download/latest/blorg-$(uname -m) \
        -o blorg && chmod +x blorg

Or install with go:

    $ go install github.com/facundoolano/blorg


Usage:

```shell
$ blorg init myblog
  site name: My Blog
  author: Facundo

  added myblog/README.md
  added myblog/.gitignore
  added myblog/config.yml
  added myblog/layouts/base.html
  added myblog/layouts/post.html
  added myblog/src/index.html
  added myblog/assets/css/main.css
  added myblog/src/blog/hello.org
  added myblog/src/feed.xml
  added myblog/src/tags.html
$ cd myblog
$ blorg post "My First Post"
  added src/blog/my-first-post.org
$ blorg serve &
  server running at http://localhost:4001/
$ cat >> src/blog/test.org <<EOF
  # Hello world!

  this is my *first* post.
  EOF
$ open http://localhost:4001/blog/my-first-post
$ blorg pub
  drafts:
    1. blog/my-first-post.org
  choose file to publish: 1
$ blorg build
  added target/index.html
  added target/assets/css/main.css
  added target/blog/hello.html
  added target/blog/my-first-post.html
  added target/feed.xml
  added target/tags.html
```

For more details see the:

  - [Tutorial](https://blorg.olano.dev#tutorial)
  - [Docs](https://blorg.olano.dev#docs)
  - [Development blog](https://blorg.olano.dev#blog)
