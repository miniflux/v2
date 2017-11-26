Go Library for Miniflux
=======================
[![Build Status](https://travis-ci.org/miniflux/miniflux-go.svg?branch=master)](https://travis-ci.org/miniflux/miniflux-go)
[![GoDoc](https://godoc.org/github.com/miniflux/miniflux-go?status.svg)](https://godoc.org/github.com/miniflux/miniflux-go)

Client library for Miniflux REST API.

Requirements
------------

- Miniflux >= 2.0.0
- Go >= 1.9

Installation
------------

```bash
go get -u github.com/miniflux/miniflux-go
```

Example
-------

```go
package main

import (
	"fmt"

	"github.com/miniflux/miniflux-go"
)

func main() {
    client := miniflux.NewClient("https://api.example.org", "admin", "secret")

    // Fetch all feeds.
    feeds, err := client.Feeds()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(feeds)
}
```

Credits
-------

- Author: Frédéric Guillot
- Distributed under MIT License
