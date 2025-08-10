influxeed-engine API Client
===================

[![PkgGoDev](https://pkg.go.dev/badge/influxeed-engine.app/v2/client)](https://pkg.go.dev/influxeed-engine.app/v2/client)

Client library for influxeed-engine REST API.

Installation
------------

```bash
go get -u influxeed-engine.app/v2/client
```

Example
-------

```go
package main

import (
	"fmt"
	"os"

	influxeed-engine "influxeed-engine.app/v2/client"
)

func main() {
    // Authentication with username/password:
    client := influxeed-engine.NewClient("https://api.example.org", "admin", "secret")

    // Authentication with an API Key:
    client := influxeed-engine.NewClient("https://api.example.org", "my-secret-token")

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

    err = os.WriteFile("opml.xml", opml, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }
}
```
