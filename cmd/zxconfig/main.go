package main

import (
	"github.com/zostay/dev-tools/cmd/zxconfig/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
