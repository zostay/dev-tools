package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-changelog/changelogImpl"
)

func main() {
	plugin.RunPlugin(&changelogImpl.Plugin{})
}
