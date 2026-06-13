package rss

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func BenchmarkParseRidiculousEntry(b *testing.B) {
	data := `<?xml version="1.0" encoding="utf-8"?>
		<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/">
		<channel>
		<title>My Example Feed</title>
		<link>https://example.org</link>
		<item>
			<title>Example Item</title>
			<link>http://www.example.org/entries/1</link>
			<enclosure type="application/x-bittorrent" url="https://example.org/file3.torrent" length="670053113">
			</enclosure>
			<media:group>
				<media:content type="application/x-bittorrent" url="https://example.org/file1.torrent"></media:content>
				<media:content type="application/x-bittorrent" url="https://example.org/file2.torrent" isDefault="true"></media:content>
				<media:content type="application/x-bittorrent" url="https://example.org/file3.torrent"></media:content>
				<media:content type="application/x-bittorrent" url="https://example.org/file4.torrent"></media:content>
				<media:content type="application/x-bittorrent" url="https://example.org/file4.torrent"></media:content>
				<media:content type="application/x-bittorrent" url=" file5.torrent  " fileSize="42"></media:content>
				<media:content type="application/x-bittorrent" url="  " fileSize="42"></media:content>
				<media:rating>nonadult</media:rating>
			</media:group>
			<media:thumbnail url="https://example.org/image.jpg" height="122" width="223"></media:thumbnail>
			<media:thumbnail url="https://example.org/thumbnail.jpg" />
			<media:thumbnail url="https://example.org/thumbnail.jpg" />
			<media:thumbnail url=" thumbnail.jpg  " />
			<media:thumbnail url="   " />
			<media:content url="https://example.org/media1.jpg" medium="image">
				<media:title type="html">Some Title for Media 1</media:title>
			</media:content>
			<media:content url="   /media2.jpg   " medium="image" />
			<media:content url="    " medium="image" />
			<media:peerLink type="application/x-bittorrent" href="https://www.example.org/file.torrent" />
			<media:peerLink type="application/x-bittorrent" href="https://www.example.org/file.torrent" />
			<media:peerLink type="application/x-bittorrent" href="  file2.torrent   " />
			<media:peerLink type="application/x-bittorrent" href="    " />
		</item>
		</channel>
		</rss>`

	var feed *model.Feed
	for b.Loop() {
		var err error
		feed, err = Parse("https://example.org/", strings.NewReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = feed
}
