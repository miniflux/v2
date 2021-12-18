# html2text

[![Documentation](https://godoc.org/github.com/jaytaylor/html2text?status.svg)](https://godoc.org/github.com/jaytaylor/html2text)
[![Build Status](https://travis-ci.org/jaytaylor/html2text.svg?branch=master)](https://travis-ci.org/jaytaylor/html2text)
[![Report Card](https://goreportcard.com/badge/github.com/jaytaylor/html2text)](https://goreportcard.com/report/github.com/jaytaylor/html2text)

### Converts HTML into text of the markdown-flavored variety

## Introduction

Ensure your emails are readable by all!

Turns HTML into raw text, useful for sending fancy HTML emails with an equivalently nicely formatted TXT document as a fallback (e.g. for people who don't allow HTML emails or have other display issues).

html2text is a simple golang package for rendering HTML into plaintext.

There are still lots of improvements to be had, but FWIW this has worked fine for my [basic] HTML-2-text needs.

It requires go 1.x or newer ;)

## Download the package

```bash
go get jaytaylor.com/html2text
```

## Example usage

### Library

```go
package main

import (
	"fmt"

	"jaytaylor.com/html2text"
)

func main() {
	inputHTML := `
<html>
  <head>
    <title>My Mega Service</title>
    <link rel=\"stylesheet\" href=\"main.css\">
    <style type=\"text/css\">body { color: #fff; }</style>
  </head>

  <body>
    <div class="logo">
      <a href="http://jaytaylor.com/"><img src="/logo-image.jpg" alt="Mega Service"/></a>
    </div>

    <h1>Welcome to your new account on my service!</h1>

    <p>
      Here is some more information:

      <ul>
        <li>Link 1: <a href="https://example.com">Example.com</a></li>
        <li>Link 2: <a href="https://example2.com">Example2.com</a></li>
        <li>Something else</li>
      </ul>
    </p>

    <table>
      <thead>
        <tr><th>Header 1</th><th>Header 2</th></tr>
      </thead>
      <tfoot>
        <tr><td>Footer 1</td><td>Footer 2</td></tr>
      </tfoot>
      <tbody>
        <tr><td>Row 1 Col 1</td><td>Row 1 Col 2</td></tr>
        <tr><td>Row 2 Col 1</td><td>Row 2 Col 2</td></tr>
      </tbody>
    </table>
  </body>
</html>`

	text, err := html2text.FromString(inputHTML, html2text.Options{PrettyTables: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
```

Output:
```
Mega Service ( http://jaytaylor.com/ )

******************************************
Welcome to your new account on my service!
******************************************

Here is some more information:

* Link 1: Example.com ( https://example.com )
* Link 2: Example2.com ( https://example2.com )
* Something else

+-------------+-------------+
|  HEADER 1   |  HEADER 2   |
+-------------+-------------+
| Row 1 Col 1 | Row 1 Col 2 |
| Row 2 Col 1 | Row 2 Col 2 |
+-------------+-------------+
|  FOOTER 1   |  FOOTER 2   |
+-------------+-------------+
```

### Command line

```
echo '<div>hi</div>' | html2text
```

## Unit-tests

Running the unit-tests is straightforward and standard:

```bash
go test
```

# License

Permissive MIT license.

## Contact

You are more than welcome to open issues and send pull requests if you find a bug or want a new feature.

If you appreciate this library please feel free to drop me a line and tell me!  It's always nice to hear from people who have benefitted from my work.

Email: jay at (my github username).com

Twitter: [@jtaylor](https://twitter.com/jtaylor)

# Alternatives

https://github.com/k3a/html2text - Lightweight
