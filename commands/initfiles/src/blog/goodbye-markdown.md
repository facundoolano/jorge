---
title: Goodbye Markdown...
tags: [blog]
date: 2024-02-16
layout: post
---

## For the record

For the record, even though it has *org* in the name, jorge can also render markdown,
thanks to [goldmark](https://github.com/yuin/goldmark/).

Let's look at some code:

``` python
import os

def hello():
  print("Hello World!")
  os.exit(0)

hello()
```

Let's try some gfm extensions. This is ~~strikethrough~~.

| foo | bar |
| --- | --- |
| baz | bim |

Would it support footnotes[^1]?

[Next time](./hello-org), I'll talk about org-mode posts.

[^1]: apparently it would?
