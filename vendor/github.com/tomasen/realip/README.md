a golang library that can get client's real public ip address from http request headers

[![Build Status](https://travis-ci.org/tomasen/realip.svg?branch=master)](https://travis-ci.org/Tomasen/realip)
[![GoDoc](https://godoc.org/github.com/Tomasen/realip?status.svg)](http://godoc.org/github.com/Tomasen/realip)


* follow the rule of X-FORWARDED-FOR/rfc7239
* follow the rule of X-Real-Ip
* lan/intranet IP address filtered

## Developing

Commited code must pass:

* [golint](https://github.com/golang/lint)
* [go vet](https://godoc.org/golang.org/x/tools/cmd/vet)
* [gofmt](https://golang.org/cmd/gofmt)
* [go test](https://golang.org/cmd/go/#hdr-Test_packages):
