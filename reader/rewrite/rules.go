// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite // import "miniflux.app/reader/rewrite"

// List of predefined rewrite rules (alphabetically sorted)
// Available rules: "add_image_title", "add_youtube_video"
// domain => rule name
var predefinedRules = map[string]string{
	"abstrusegoose.com":      "add_image_title",
	"amazingsuperpowers.com": "add_image_title",
	"cowbirdsinlove.com":     "add_image_title",
	"drawingboardcomic.com":  "add_image_title",
	"exocomics.com":          "add_image_title",
	"happletea.com":          "add_image_title",
	"imogenquest.net":        "add_image_title",
	"lukesurl.com":           "add_image_title",
	"mercworks.net":          "add_image_title",
	"mrlovenstein.com":       "add_image_title",
	"nedroid.com":            "add_image_title",
	"oglaf.com":              "add_image_title",
	"optipess.com":           "add_image_title",
	"peebleslab.com":         "add_image_title",
	"sentfromthemoon.com":    "add_image_title",
	"thedoghousediaries.com": "add_image_title",
	"treelobsters.com":       "add_image_title",
	"youtube.com":            "add_youtube_video",
	"xkcd.com":               "add_image_title",
}
