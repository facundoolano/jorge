# blorg
A presonal (small + opinionated) site generator with org-mode support.

(NOTE: this is stil a WIP, the doc below is a wishlist, not the current behavior.)

Install from binary:

    $ wget https://github.com/facundoolano/blorg/releases/download/latest/blorg-$(uname -m) \
        -o blorg && chmod +x blorg

Or install with go:

    $ go install github.com/facundoolano/blorg


Usage:

```bash
$ blorg init myblog
> site name: My Blog
> author: Facundo

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
  added draft src/blog/my-first-post.org

# serve the site locally with live reload
$ blorg serve
  server running at http://localhost:4001/

# browse to the new post
$ open http://localhost:4001/blog/my-first-post

# add some content
$ cat >> src/blog/test.org <<EOF
# Hello world!

this is my *first* post.
EOF

# remove the draft flag before publishing
$ sed -i '/^draft: true$/d' src/blog/my-first-post.org

$ blorg build
  wrote target/index.html
  wrote target/assets/css/main.css
  wrote target/blog/hello.html
  wrote target/blog/my-first-post.html
  wrote target/feed.xml
  wrote target/tags.html
```

For more details see the:

  - [Tutorial](https://blorg.olano.dev#tutorial)
  - [Docs](https://blorg.olano.dev#docs)
  - [Development blog](https://blorg.olano.dev#blog)
