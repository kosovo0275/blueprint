package bootstrap

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/blueprint"
)

func bootstrapVariable(name string, value func() string) blueprint.Variable {
	return pctx.VariableFunc(name, func(config interface{}) (string, error) {
		return value(), nil
	})
}

var (
	// These variables are the only configuration
	// needed by the boostrap modules.
	srcDir = bootstrapVariable("srcDir", func() string {
		return SrcDir
	})
	buildDir = bootstrapVariable("buildDir", func() string {
		return BuildDir
	})
	ninjaBuildDir = bootstrapVariable("ninjaBuildDir", func() string {
		return NinjaBuildDir
	})
	goRoot = bootstrapVariable("goRoot", func() string {
		goroot := runtime.GOROOT()
		// Prefer to omit absolute paths from the ninja file
		if cwd, err := os.Getwd(); err == nil {
			if relpath, err := filepath.Rel(cwd, goroot); err == nil {
				if !strings.HasPrefix(relpath, "../") {
					goroot = relpath
				}
			}
		}
		return goroot
	})
	compileCmd = bootstrapVariable("compileCmd", func() string {
		return "$goRoot/pkg/tool/" + runtime.GOOS + "_" + runtime.GOARCH + "/compile"
	})
	linkCmd = bootstrapVariable("linkCmd", func() string {
		return "$goRoot/pkg/tool/" + runtime.GOOS + "_" + runtime.GOARCH + "/link"
	})
)

type ConfigInterface interface {
	// GeneratingPrimaryBuilder should return true if
	// this build invocation is creating a
	// .bootstrap/build.ninja file to be used to
	// build the primary builder
	GeneratingPrimaryBuilder() bool
}

type ConfigRemoveAbandonedFilesUnder interface {
	// RemoveAbandonedFilesUnder should return a slice
	// of path prefixes that will be cleaned of files
	// that are no longer active targets, but are
	// listed in the .ninja_log.
	RemoveAbandonedFilesUnder() []string
}

type ConfigBlueprintToolLocation interface {
	// BlueprintToolLocation can return a path name
	// to install blueprint tools designed for end
	// users (bpfmt, bpmodify, and anything else using
	// blueprint_go_binary).
	BlueprintToolLocation() string
}

type StopBefore int

const (
	StopBeforePrepareBuildActions StopBefore = 1
)

type ConfigStopBefore interface {
	StopBefore() StopBefore
}

type Stage int

const (
	StagePrimary Stage = iota
	StageMain
)

type Config struct {
	stage Stage

	topLevelBlueprintsFile string

	runGoTests     bool
	moduleListFile string
}
