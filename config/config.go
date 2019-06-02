// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

// Opts contains configuration options after parsing.
var Opts *Options

// ParseConfig parses configuration options.
func ParseConfig() (err error) {
	Opts, err = parse()
	return err
}
