# getlang

[![GoDoc](https://godoc.org/github.com/rylans/getlang?status.svg)](https://godoc.org/github.com/rylans/getlang) [![Go Report Card](https://goreportcard.com/badge/github.com/rylans/getlang)](https://goreportcard.com/report/github.com/rylans/getlang) [![Build Status](https://travis-ci.org/rylans/getlang.svg?branch=master)](https://travis-ci.org/rylans/getlang) ![cover.run go](https://cover.run/go/github.com/rylans/getlang.svg?tag=golang-1.10)

getlang provides fast natural language detection in Go.

## Features

* Offline -- no internet connection required
* Supports [29 languages](https://github.com/rylans/getlang/blob/master/LANGUAGES.md)
* Provides ISO 639 language codes
* Fast

## Getting started

Installation:
```sh
    go get -u github.com/rylans/getlang
```

example:
```go
package main

import (
	"fmt"
	"github.com/rylans/getlang"
)

func main(){
  info := getlang.FromString("Wszyscy ludzie rodzą się wolni i równi w swojej godności i prawach")
  fmt.Println(info.LanguageCode(), info.Confidence())
}
```

## Documentation
[getlang on godoc](https://godoc.org/github.com/rylans/getlang)

## License
[MIT](https://github.com/rylans/getlang/blob/master/LICENSE)

## Acknowledgements and Citations
* Thanks to [abadojack](https://github.com/abadojack) for the trigram generation logic in whatlanggo
* Cavnar, William B., and John M. Trenkle. "N-gram-based text categorization." Ann arbor mi 48113.2 (1994): 161-175.
