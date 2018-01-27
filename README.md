Miniflux 2
==========
[![Build Status](https://travis-ci.org/miniflux/miniflux.svg?branch=master)](https://travis-ci.org/miniflux/miniflux)
[![GoDoc](https://godoc.org/github.com/miniflux/miniflux?status.svg)](https://godoc.org/github.com/miniflux/miniflux)
[![Documentation Status](https://readthedocs.org/projects/miniflux/badge/?version=latest)](https://docs.miniflux.net/)

Miniflux is a minimalist and opinionated feed reader:

- Written in Go (Golang)
- Works only with Postgresql
- Doesn't use any ORM
- Doesn't use any complicated framework
- Use only modern vanilla Javascript (ES6 and Fetch API)
- The number of features is voluntarily limited

It's simple, fast, lightweight and super easy to install.

Miniflux 2 is a rewrite of [Miniflux 1.x](https://github.com/miniflux/miniflux-legacy) in Golang.

Documentation
-------------

The Miniflux documentation is available here: <https://docs.miniflux.net/>

- [Opinionated?](https://docs.miniflux.net/en/latest/opinionated.html)
- [Features](https://docs.miniflux.net/en/latest/features.html)
- [Requirements](https://docs.miniflux.net/en/latest/requirements.html)
- [Installation](https://docs.miniflux.net/en/latest/installation.html)
- [Upgrading to a new version](https://docs.miniflux.net/en/latest/upgrade.html)
- [Configuration](https://docs.miniflux.net/en/latest/configuration.html)

Running local (development) instance
------------------------------------

#### First create a postgres instance:

```
docker run --name miniflux_db -p 5432:5432 -d -e POSTGRES_USER=miniflux -e POSTGRES_PASSWORD=secret postgres
```

#### Then run your local version:

```
DATABASE_URL=postgres://miniflux:secret@localhost/miniflux?sslmode=disable go run main.go -migrate
DATABASE_URL=postgres://miniflux:secret@localhost/miniflux?sslmode=disable go run main.go -create-admin
DATABASE_URL=postgres://miniflux:secret@localhost/miniflux?sslmode=disable go run main.go
```

If everything went well, then connect to http://localhost:8080

Credits
-------

- Author: Frédéric Guillot
- Distributed under Apache 2.0 License
