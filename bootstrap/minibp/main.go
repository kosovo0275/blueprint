package main

import (
	"flag"
	"path/filepath"

	"github.com/google/blueprint"
	"github.com/google/blueprint/bootstrap"
)

var runAsPrimaryBuilder bool
var buildPrimaryBuilder bool

func init() {
	flag.BoolVar(&runAsPrimaryBuilder, "p", false, "run as a primary builder")
}

type Config struct {
	generatingPrimaryBuilder bool
}

func (c Config) GeneratingPrimaryBuilder() bool {
	return c.generatingPrimaryBuilder
}

func (c Config) RemoveAbandonedFilesUnder() []string {
	return []string{filepath.Join(bootstrap.BuildDir, ".bootstrap")}
}

func main() {
	flag.Parse()

	ctx := blueprint.NewContext()
	if !runAsPrimaryBuilder {
		ctx.SetIgnoreUnknownModuleTypes(true)
	}

	config := Config{
		generatingPrimaryBuilder: !runAsPrimaryBuilder,
	}

	bootstrap.Main(ctx, config)
}
