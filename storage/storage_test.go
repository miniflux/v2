package storage

import (
	"testing"
	"unicode/utf8"
)

func TestStorage_LogKeywordForContent(t *testing.T) {
	if len(ReplacePunctuation(" ")) > 0 {
		t.Error("must trim space")
	}

	if len(ReplacePunctuation("，")) > 0 {
		t.Error("must trim Punctuation")
	}

	v := ReplacePunctuation("，你,|")
	if utf8.RuneCountInString(v) != 1 {
		t.Error("must keep char")
	}

	if ReplacePunctuation("，你,|好。世界") != "你好世界" {
		t.Error("must eq char")
	}

	if ReplacePunctuation("，h,|ello。world") != "helloworld" {
		t.Error("must eq char")
	}
}
