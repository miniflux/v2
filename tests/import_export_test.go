// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package tests

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestExport(t *testing.T) {
	client := createClient(t)

	output, err := client.Export()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(output), "<?xml") {
		t.Fatalf(`Invalid OPML export, got "%s"`, string(output))
	}
}

func TestImport(t *testing.T) {
	client := createClient(t)

	data := `<?xml version="1.0" encoding="UTF-8"?>
    <opml version="2.0">
        <body>
            <outline text="Test Category">
				<outline title="Test" text="Test" xmlUrl="` + testFeedURL + `" htmlUrl="` + testWebsiteURL + `"></outline>
			</outline>
		</body>
	</opml>`

	b := bytes.NewReader([]byte(data))
	err := client.Import(io.NopCloser(b))
	if err != nil {
		t.Fatal(err)
	}
}
