package gojieba

import (
	"github.com/yanyiwu/gojieba/deps/cppjieba"
	"github.com/yanyiwu/gojieba/deps/limonp"
	"github.com/yanyiwu/gojieba/dict"
)

func init() {
	dict.Init()
	limonp.Init()
	cppjieba.Init()
}
