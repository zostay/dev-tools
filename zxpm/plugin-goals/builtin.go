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
	goalRequest  = "request"
	goalTest     = "test"
)

func describeBuild() *GoalDescription {
	return &GoalDescription{
		name:   goalBuild,
		plugin: pluginFullName,
		short:  "Syntax check and prepare for development.",
	}
}

func describeDeploy() *GoalDescription {
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

func describeInstall() *GoalDescription {
	return &GoalDescription{
		name:   goalInstall,
		plugin: pluginFullName,
		short:  "Install software and assets locally.",
	}
}

func describeLint() *GoalDescription {
	return &GoalDescription{
		name:   "lint",
		plugin: pluginFullName,
		short:  "Check files and data for errors and anti-patterns.",
		alias:  []string{"analyze"},
	}
}

func describeRequest() *GoalDescription {
	return &GoalDescription{
		name:   "request",
		plugin: pluginFullName,
		short:  "Request the merger of a code patch.",
		alias:  []string{"pull-request", "pr", "merge-request", "mr"},
	}
}

func describeRelease() *GoalDescription {
	return &GoalDescription{
		name:   "release",
		plugin: pluginFullName,
		short:  "Mint and publish a release.",
	}
}

func describeTest() *GoalDescription {
	return &GoalDescription{
		name:   "test",
		plugin: pluginFullName,
		short:  "Run tests.",
	}
}
