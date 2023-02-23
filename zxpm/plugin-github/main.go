package main

import (
	"github.com/zostay/dev-tools/zxpm/plugin-github/githubImpl"
	"github.com/zostay/dev-tools/zxpm/plugin/metal"
)

func main() {
	metal.RunPlugin(&githubImpl.Plugin{})
}
