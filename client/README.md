Miniflux API Client
===================

Client library for Miniflux REST API.

Installation
------------

```bash
go get -u miniflux.app/client
```

Example
-------

```go
package main

import (
	"fmt"
    "io/ioutil"

	miniflux "miniflux.app/client"
)

func main() {
    client := miniflux.New("https://api.example.org", "admin", "secret")

    // Fetch all feeds.
    feeds, err := client.Feeds()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(feeds)

    // Backup your feeds to an OPML file.
    opml, err := client.Export()
    if err != nil {
        fmt.Println(err)
        return
    }

    err = ioutil.WriteFile("opml.xml", opml, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }
}
```
