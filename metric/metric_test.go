package metric

import "testing"

func TestLogKeywordForContent(t *testing.T) {
	LogKeywordForContent("")
	LogKeywordForContent("<p>我来到北京清华大学</p>")
	LogKeywordForContent("我来到北京清华大学")
}
