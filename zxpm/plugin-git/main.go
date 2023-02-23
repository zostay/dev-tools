package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-git/gitImpl"
)

func main() {
	plugin.RunPlugin(&gitImpl.Plugin{})
}
