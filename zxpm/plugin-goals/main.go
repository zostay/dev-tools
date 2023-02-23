package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin-goals/goalsImpl"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
)

func main() {
	metal.RunPlugin(&goalsImpl.Plugin{})
}
