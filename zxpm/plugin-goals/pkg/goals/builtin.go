package goals

type here struct{}

const pluginFullName = "zxpm-plugin-goals"

const (
	NameBuild    = "build"
	NameDeploy   = "deploy"
	NameGenerate = "generate"
	NameInfo     = "info"
	NameInit     = "init"
	NameInstall  = "install"
	NameLint     = "lint"
	NameRelease  = "release"
	NameRequest  = "request"
	NameTest     = "test"
)

func DescribeBuild() *GoalDescription {
	return &GoalDescription{
		name:   NameBuild,
		plugin: pluginFullName,
		short:  "Syntax check and prepare for development.",
	}
}

func DescribeDeploy() *GoalDescription {
	return &GoalDescription{
		name:   NameDeploy,
		plugin: pluginFullName,
		short:  "Deploy software to a remote server.",
	}
}

func DescribeGenerate() *GoalDescription {
	return &GoalDescription{
		name:   NameGenerate,
		plugin: pluginFullName,
		short:  "Perform code generation tasks.",
	}
}

func DescribeInfo() *GoalDescription {
	return &GoalDescription{
		name:   NameInfo,
		plugin: pluginFullName,
		short:  "Describe information about the project.",
	}
}

func DescribeInit() *GoalDescription {
	return &GoalDescription{
		name:   NameInit,
		plugin: pluginFullName,
		short:  "Initialize a new project directory.",
	}
}

func DescribeInstall() *GoalDescription {
	return &GoalDescription{
		name:   NameInstall,
		plugin: pluginFullName,
		short:  "Install software and assets locally.",
	}
}

func DescribeLint() *GoalDescription {
	return &GoalDescription{
		name:    NameLint,
		plugin:  pluginFullName,
		short:   "Check files and data for errors and anti-patterns.",
		aliases: []string{"analyze"},
	}
}

func DescribeRequest() *GoalDescription {
	return &GoalDescription{
		name:    NameRequest,
		plugin:  pluginFullName,
		short:   "Request the merger of a code patch.",
		aliases: []string{"pull-request", "pr", "merge-request", "mr"},
	}
}

func DescribeRelease() *GoalDescription {
	return &GoalDescription{
		name:   NameRelease,
		plugin: pluginFullName,
		short:  "Mint and publish a release.",
	}
}

func DescribeTest() *GoalDescription {
	return &GoalDescription{
		name:   NameTest,
		plugin: pluginFullName,
		short:  "Run tests.",
	}
}
