package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-goals/goalsImpl"
)

func main() {
	plugin.RunPlugin(&goalsImpl.Plugin{})
}
