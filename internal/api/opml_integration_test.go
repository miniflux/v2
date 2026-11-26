// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	miniflux "miniflux.app/v2/client"
)

func TestExportEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	if _, err := regularUserClient.CreateFeed(&miniflux.FeedCreationRequest{FeedURL: fixture.feedURL}); err != nil {
		t.Fatal(err)
	}

	exportedData, err := regularUserClient.Export()
	if err != nil {
		t.Fatal(err)
	}

	if len(exportedData) == 0 {
		t.Fatalf(`Invalid exported data, got empty`)
	}

	if !strings.HasPrefix(string(exportedData), "<?xml") {
		t.Fatalf(`Invalid OPML export, got %q`, string(exportedData))
	}
}

func TestImportEndpoint(t *testing.T) {
	t.Parallel()

	testConfig := requireIntegrationTestConfig(t)
	fixture := newFeedFixture(t)

	regularUserClient, _ := testConfig.newRegularUser(t)

	data := `<?xml version="1.0" encoding="UTF-8"?>
    <opml version="2.0">
        <body>
            <outline text="Test Category">
				<outline title="Test" text="Test" xmlUrl="` + fixture.feedURL + `" htmlUrl="` + fixture.websiteURL + `"></outline>
			</outline>
		</body>
	</opml>`

	bytesReader := bytes.NewReader([]byte(data))
	if err := regularUserClient.Import(io.NopCloser(bytesReader)); err != nil {
		t.Fatal(err)
	}
}
