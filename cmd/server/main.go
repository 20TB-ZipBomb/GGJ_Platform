package main

import (
	"flag"

	"github.com/20TB-ZipBomb/GGJ_Platform/pkg/app"
)

var addr = flag.String("addr", "localhost:4041", "http service address")

func main() {
	app.Run()
}
