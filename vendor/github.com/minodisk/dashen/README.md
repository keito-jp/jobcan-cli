# dashen [![Codeship Status for minodisk/dashen](https://img.shields.io/codeship/8c4d91c0-ecf4-0134-e2b6-0a42fa094665/master.svg?style=flat)](https://app.codeship.com/projects/208467) [![Go Report Card](https://goreportcard.com/badge/github.com/minodisk/dashen)](https://goreportcard.com/report/github.com/minodisk/dashen) [![codecov](https://codecov.io/gh/minodisk/dashen/branch/master/graph/badge.svg)](https://codecov.io/gh/minodisk/dashen) [![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat)](https://godoc.org/github.com/minodisk/dashen) [![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Listen the Amazon Dash Button is pressed in the intranet.

## Usage

```sh
go get -u github.com/minodisk/dashen
```

```go
package main

import (
	"log"

	"github.com/minodisk/dashen"
)

func main() {
	d := dashen.New()
	d.Subscribe("00:00:00:00:00:00", func() {
		log.Println("detected")
	})
	if err := d.Listen(); err != nil {
		panic(err)
	}
}
```
