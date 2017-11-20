# JSON [![GoDoc](http://godoc.org/github.com/tdewolff/parse/json?status.svg)](http://godoc.org/github.com/tdewolff/parse/json) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/parse/json)](http://gocover.io/github.com/tdewolff/parse/json)

This package is a JSON lexer (ECMA-404) written in [Go][1]. It follows the specification at [JSON](http://json.org/). The lexer takes an io.Reader and converts it into tokens until the EOF.

## Installation
Run the following command

	go get github.com/tdewolff/parse/json

or add the following import and run project with `go get`

	import "github.com/tdewolff/parse/json"

## Parser
### Usage
The following initializes a new Parser with io.Reader `r`:
``` go
p := json.NewParser(r)
```

To tokenize until EOF an error, use:
``` go
for {
	gt, text := p.Next()
	switch gt {
	case json.ErrorGrammar:
		// error or EOF set in p.Err()
		return
	// ...
	}
}
```

All grammars:
``` go
ErrorGrammar       GrammarType = iota // extra grammar when errors occur
WhitespaceGrammar                     // space \t \r \n
LiteralGrammar                        // null true false
NumberGrammar
StringGrammar
StartObjectGrammar // {
EndObjectGrammar   // }
StartArrayGrammar  // [
EndArrayGrammar    // ]
```

### Examples
``` go
package main

import (
	"os"

	"github.com/tdewolff/parse/json"
)

// Tokenize JSON from stdin.
func main() {
	p := json.NewParser(os.Stdin)
	for {
		gt, text := p.Next()
		switch gt {
		case json.ErrorGrammar:
			if p.Err() != io.EOF {
				fmt.Println("Error on line", p.Line(), ":", p.Err())
			}
			return
		case json.LiteralGrammar:
			fmt.Println("Literal", string(text))
		case json.NumberGrammar:
			fmt.Println("Number", string(text))
		// ...
		}
	}
}
```

## License
Released under the [MIT license](https://github.com/tdewolff/parse/blob/master/LICENSE.md).

[1]: http://golang.org/ "Go Language"
