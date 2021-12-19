# 结巴分词 Go 语言版：Jiebago

[![Go](https://github.com/Soontao/jiebago/actions/workflows/go.yml/badge.svg)](https://github.com/Soontao/jiebago/actions/workflows/go.yml)

[结巴分词](https://github.com/fxsjy/jieba) 是由 [@fxsjy](https://github.com/fxsjy) 使用 Python 编写的中文分词组件，jiebago 是结巴分词的 Golang 语言实现。


## 安装

```
go get github.com/Soontao/jiebago/...
```

## 使用

```go
package main

import (
        "fmt"

        "github.com/Soontao/jiebago"
)

var seg jiebago.Segmenter

func init() {
        seg.LoadDictionary("dict.txt")
}

func print(ch <-chan string) {
        for word := range ch {
                fmt.Printf(" %s /", word)
        }
        fmt.Println()
}

func Example() {
        fmt.Print("【全模式】：")
        print(seg.CutAll("我来到北京清华大学"))

        fmt.Print("【精确模式】：")
        print(seg.Cut("我来到北京清华大学", false))

        fmt.Print("【新词识别】：")
        print(seg.Cut("他来到了网易杭研大厦", true))

        fmt.Print("【搜索引擎模式】：")
        print(seg.CutForSearch("小明硕士毕业于中国科学院计算所，后在日本京都大学深造", true))
}
```

输出结果：

```bash
【全模式】： 我 / 来到 / 北京 / 清华 / 清华大学 / 华大 / 大学 /

【精确模式】： 我 / 来到 / 北京 / 清华大学 /

【新词识别】： 他 / 来到 / 了 / 网易 / 杭研 / 大厦 /

【搜索引擎模式】： 小明 / 硕士 / 毕业 / 于 / 中国 / 科学 / 学院 / 科学院 / 中国科学院 / 计算 / 计算所 / ， / 后 / 在 / 日本 / 京都 / 大学 / 日本京都大学 / 深造 /
```

更多信息请参考[文档](https://godoc.org/github.com/Soontao/jiebago)。

## 分词速度

- 2MB / Second in Full Mode
- 700KB / Second in Default Mode
- Test Env: AMD Phenom(tm) II X6 1055T CPU @ 2.8GHz; 《金庸全集》 

## 许可证

MIT: http://wangbin.mit-license.org
