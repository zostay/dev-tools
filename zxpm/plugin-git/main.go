package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin-git/gitImpl"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
)

func main() {
	metal.RunPlugin(&gitImpl.Plugin{})
}
