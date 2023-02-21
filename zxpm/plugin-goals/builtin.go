package plugin_goals

import (
	"reflect"
)

type here struct{}

var pluginFullName = reflect.TypeOf(here{}).PkgPath()

const (
	goalBuild    = "build"
	goalDeploy   = "deploy"
	goalGenerate = "generate"
	goalInstall  = "install"
	goalLint     = "lint"
	goalRelease  = "release"
	goalRequest  = "request"
	goalTest     = "test"
)

func DescribeBuild() *GoalDescription {
	return &GoalDescription{
		name:   goalBuild,
		plugin: pluginFullName,
		short:  "Syntax check and prepare for development.",
	}
}

func DescribeDeploy() *GoalDescription {
	return &GoalDescription{
		name:   goalDeploy,
		plugin: pluginFullName,
		short:  "Deploy software to a remote server.",
	}
}

func describeGenerate() *GoalDescription {
	return &GoalDescription{
		name:   goalGenerate,
		plugin: pluginFullName,
		short:  "Perform code generation tasks.",
	}
}

func DescribeInstall() *GoalDescription {
	return &GoalDescription{
		name:   goalInstall,
		plugin: pluginFullName,
		short:  "Install software and assets locally.",
	}
}

func DescribeLint() *GoalDescription {
	return &GoalDescription{
		name:   goalLint,
		plugin: pluginFullName,
		short:  "Check files and data for errors and anti-patterns.",
		alias:  []string{"analyze"},
	}
}

func DescribeRequest() *GoalDescription {
	return &GoalDescription{
		name:   goalRequest,
		plugin: pluginFullName,
		short:  "Request the merger of a code patch.",
		alias:  []string{"pull-request", "pr", "merge-request", "mr"},
	}
}

func DescribeRelease() *GoalDescription {
	return &GoalDescription{
		name:   goalRelease,
		plugin: pluginFullName,
		short:  "Mint and publish a release.",
	}
}

func DescribeTest() *GoalDescription {
	return &GoalDescription{
		name:   goalTest,
		plugin: pluginFullName,
		short:  "Run tests.",
	}
}
