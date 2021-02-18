// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main // import "miniflux.app"

//go:generate go run generate.go
//go:generate gofmt -s -w ui/static/js.go
//go:generate gofmt -s -w template/views.go
//go:generate gofmt -s -w template/common.go

import (
	"miniflux.app/cli"
)

func main() {
	cli.Parse()
}
