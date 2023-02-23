package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin-changelog/changelogImpl"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
)

func main() {
	metal.RunPlugin(&changelogImpl.Plugin{})
}
