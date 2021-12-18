# GoJieba [English](README_EN.md)

[![Build Status](https://travis-ci.org/yanyiwu/gojieba.png?branch=master)](https://travis-ci.org/yanyiwu/gojieba) 
[![Author](https://img.shields.io/badge/author-@yanyiwu-blue.svg?style=flat)](http://yanyiwu.com/) 
[![Donate](https://img.shields.io/badge/donate-eos_git@yanyiwu-orange.svg)](https://eospark.com/account/gitatyanyiwu)
[![Tag](https://img.shields.io/github/v/tag/yanyiwu/gojieba.svg)](https://github.com/yanyiwu/gojieba/releases)
[![Performance](https://img.shields.io/badge/performance-excellent-brightgreen.svg?style=flat)](http://yanyiwu.com/work/2015/06/14/jieba-series-performance-test.html) 
[![License](https://img.shields.io/badge/license-MIT-yellow.svg?style=flat)](http://yanyiwu.mit-license.org)
[![GoDoc](https://godoc.org/github.com/yanyiwu/gojieba?status.svg)](https://godoc.org/github.com/yanyiwu/gojieba)
[![Coverage Status](https://coveralls.io/repos/yanyiwu/gojieba/badge.svg?branch=master&service=github)](https://coveralls.io/github/yanyiwu/gojieba?branch=master)
[![codebeat badge](https://codebeat.co/badges/a336d042-3583-4212-8204-88da4407438e)](https://codebeat.co/projects/github-com-yanyiwu-gojieba)
[![Go Report Card](https://goreportcard.com/badge/yanyiwu/gojieba)](https://goreportcard.com/report/yanyiwu/gojieba)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go) 

[![logo](http://images.yanyiwu.com/GoJieBaLogo-v2.png)](http://yanyiwu.com/work/2015/09/14/c-cpp-go-mix-programming.html)

[GoJieba]是"结巴"中文分词的Golang语言版本。

## 简介

+ 支持多种分词方式，包括: 最大概率模式, HMM新词发现模式, 搜索引擎模式, 全模式
+ 核心算法底层由C++实现，性能高效。
+ 字典路径可配置，NewJieba(...string), NewExtractor(...string) 可变形参，当参数为空时使用默认词典(推荐方式)

## 用法

```
go get github.com/yanyiwu/gojieba
```

分词示例

```
package main

import (
	"fmt"
	"strings"

	"github.com/yanyiwu/gojieba"
)

func main() {
	var s string
	var words []string
	use_hmm := true
	x := gojieba.NewJieba()
	defer x.Free()

	s = "我来到北京清华大学"
	words = x.CutAll(s)
	fmt.Println(s)
	fmt.Println("全模式:", strings.Join(words, "/"))

	words = x.Cut(s, use_hmm)
	fmt.Println(s)
	fmt.Println("精确模式:", strings.Join(words, "/"))
	s = "比特币"
	words = x.Cut(s, use_hmm)
	fmt.Println(s)
	fmt.Println("精确模式:", strings.Join(words, "/"))

	x.AddWord("比特币")
	s = "比特币"
	words = x.Cut(s, use_hmm)
	fmt.Println(s)
	fmt.Println("添加词典后,精确模式:", strings.Join(words, "/"))

	s = "他来到了网易杭研大厦"
	words = x.Cut(s, use_hmm)
	fmt.Println(s)
	fmt.Println("新词识别:", strings.Join(words, "/"))

	s = "小明硕士毕业于中国科学院计算所，后在日本京都大学深造"
	words = x.CutForSearch(s, use_hmm)
	fmt.Println(s)
	fmt.Println("搜索引擎模式:", strings.Join(words, "/"))

	s = "长春市长春药店"
	words = x.Tag(s)
	fmt.Println(s)
	fmt.Println("词性标注:", strings.Join(words, ","))

	s = "区块链"
	words = x.Tag(s)
	fmt.Println(s)
	fmt.Println("词性标注:", strings.Join(words, ","))

	s = "长江大桥"
	words = x.CutForSearch(s, !use_hmm)
	fmt.Println(s)
	fmt.Println("搜索引擎模式:", strings.Join(words, "/"))

	wordinfos := x.Tokenize(s, gojieba.SearchMode, !use_hmm)
	fmt.Println(s)
	fmt.Println("Tokenize:(搜索引擎模式)", wordinfos)

	wordinfos = x.Tokenize(s, gojieba.DefaultMode, !use_hmm)
	fmt.Println(s)
	fmt.Println("Tokenize:(默认模式)", wordinfos)

	keywords := x.ExtractWithWeight(s, 5)
	fmt.Println("Extract:", keywords)
}
```

```
我来到北京清华大学
全模式: 我/来到/北京/清华/清华大学/华大/大学
我来到北京清华大学
精确模式: 我/来到/北京/清华大学
比特币
精确模式: 比特/币
比特币
添加词典后,精确模式: 比特币
他来到了网易杭研大厦
新词识别: 他/来到/了/网易/杭研/大厦
小明硕士毕业于中国科学院计算所，后在日本京都大学深造
搜索引擎模式: 小明/硕士/毕业/于/中国/科学/学院/科学院/中国科学院/计算/计算所/，/后/在/日本/京都/大学/日本京都大学/深造
长春市长春药店
词性标注: 长春市/ns,长春/ns,药店/n
区块链
词性标注: 区块链/nz
长江大桥
搜索引擎模式: 长江/大桥/长江大桥
长江大桥
Tokenize: [{长江 0 6} {大桥 6 12} {长江大桥 0 12}]
```

See example in [jieba_test](jieba_test.go), [extractor_test](extractor_test.go)

## Benchmark

[Jieba中文分词系列性能评测]

Unittest

```
go test ./...
```

Benchmark

```
go test -bench "Jieba" -test.benchtime 10s
go test -bench "Extractor" -test.benchtime 10s
```

## Contributors

### Code Contributors

This project exists thanks to all the people who contribute.
<a href="https://github.com/yanyiwu/gojieba/graphs/contributors"><img src="https://opencollective.com/gojieba/contributors.svg?width=890&button=false" /></a>

## Contact

+ Email: `i@yanyiwu.com`

[CppJieba]:http://github.com/yanyiwu/cppjieba
[GoJieba]:http://github.com/yanyiwu/gojieba
[Jieba]:https://github.com/fxsjy/jieba
[Jieba中文分词系列性能评测]:http://yanyiwu.com/work/2015/06/14/jieba-series-performance-test.html
