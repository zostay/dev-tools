package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/plugin-github/githubImpl"
)

func main() {
	plugin.RunPlugin(&githubImpl.Plugin{})
}
