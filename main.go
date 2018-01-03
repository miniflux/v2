// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

//go:generate go run generate.go
//go:generate gofmt -s -w sql/sql.go
//go:generate gofmt -s -w ui/static/css.go
//go:generate gofmt -s -w ui/static/bin.go
//go:generate gofmt -s -w ui/static/js.go
//go:generate gofmt -s -w template/views.go
//go:generate gofmt -s -w template/common.go
//go:generate gofmt -s -w locale/translations.go

import (
	_ "github.com/lib/pq"
	"github.com/miniflux/miniflux/cli"
)

func main() {
	cli.Parse()
}
