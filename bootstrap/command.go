package bootstrap

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"

	"github.com/google/blueprint"
	"github.com/google/blueprint/deptools"
)

var (
	outFile        string
	depFile        string
	docFile        string
	cpuprofile     string
	memprofile     string
	traceFile      string
	runGoTests     bool
	noGC           bool
	moduleListFile string

	BuildDir      string
	NinjaBuildDir string
	SrcDir        string
)

func init() {
	flag.StringVar(&outFile, "o", "build.ninja", "the Ninja file to output")
	flag.StringVar(&BuildDir, "b", ".", "the build output directory")
	flag.StringVar(&NinjaBuildDir, "n", "", "the ninja builddir directory")
	flag.StringVar(&depFile, "d", "", "the dependency file to output")
	flag.StringVar(&docFile, "docs", "", "build documentation file to output")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&traceFile, "trace", "", "write trace to file")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to file")
	flag.BoolVar(&noGC, "nogc", false, "turn off GC for debugging")
	flag.BoolVar(&runGoTests, "t", false, "build and run go tests during bootstrap")
	flag.StringVar(&moduleListFile, "l", "", "file that lists filepaths to parse")
}

func Main(ctx *blueprint.Context, config interface{}, extraNinjaFileDeps ...string) {
	if !flag.Parsed() {
		flag.Parse()
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if noGC {
		debug.SetGCPercent(-1)
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			fatalf("error opening cpuprofile: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	if traceFile != "" {
		f, err := os.Create(traceFile)
		if err != nil {
			fatalf("error opening trace: %s", err)
		}
		trace.Start(f)
		defer f.Close()
		defer trace.Stop()
	}

	if flag.NArg() != 1 {
		fatalf("no Blueprints file specified")
	}

	SrcDir = filepath.Dir(flag.Arg(0))
	if moduleListFile != "" {
		ctx.SetModuleListFile(moduleListFile)
		extraNinjaFileDeps = append(extraNinjaFileDeps, moduleListFile)
	} else {
		fatalf("-l <moduleListFile> is required and must be nonempty")
	}
	filesToParse, err := ctx.ListModulePaths(SrcDir)
	if err != nil {
		fatalf("could not enumerate files: %v\n", err.Error())
	}

	if NinjaBuildDir == "" {
		NinjaBuildDir = BuildDir
	}

	stage := StageMain
	if c, ok := config.(ConfigInterface); ok {
		if c.GeneratingPrimaryBuilder() {
			stage = StagePrimary
		}
	}

	bootstrapConfig := &Config{
		stage: stage,
		topLevelBlueprintsFile: flag.Arg(0),
		runGoTests:             runGoTests,
		moduleListFile:         moduleListFile,
	}

	ctx.RegisterBottomUpMutator("bootstrap_plugin_deps", pluginDeps)
	ctx.RegisterModuleType("bootstrap_go_package", newGoPackageModuleFactory(bootstrapConfig))
	ctx.RegisterModuleType("bootstrap_go_binary", newGoBinaryModuleFactory(bootstrapConfig, false))
	ctx.RegisterModuleType("blueprint_go_binary", newGoBinaryModuleFactory(bootstrapConfig, true))
	ctx.RegisterSingletonType("bootstrap", newSingletonFactory(bootstrapConfig))

	ctx.RegisterSingletonType("glob", globSingletonFactory(ctx))

	deps, errs := ctx.ParseFileList(filepath.Dir(bootstrapConfig.topLevelBlueprintsFile), filesToParse)
	if len(errs) > 0 {
		fatalErrors(errs)
	}

	// Add extra ninja file dependencies
	deps = append(deps, extraNinjaFileDeps...)

	extraDeps, errs := ctx.ResolveDependencies(config)
	if len(errs) > 0 {
		fatalErrors(errs)
	}
	deps = append(deps, extraDeps...)

	if docFile != "" {
		err := writeDocs(ctx, docFile)
		if err != nil {
			fatalErrors([]error{err})
		}
		return
	}

	if c, ok := config.(ConfigStopBefore); ok {
		if c.StopBefore() == StopBeforePrepareBuildActions {
			return
		}
	}

	extraDeps, errs = ctx.PrepareBuildActions(config)
	if len(errs) > 0 {
		fatalErrors(errs)
	}
	deps = append(deps, extraDeps...)

	buf := bytes.NewBuffer(nil)
	err = ctx.WriteBuildFile(buf)
	if err != nil {
		fatalf("error generating Ninja file contents: %s", err)
	}

	const outFilePermissions = 0666
	err = ioutil.WriteFile(outFile, buf.Bytes(), outFilePermissions)
	if err != nil {
		fatalf("error writing %s: %s", outFile, err)
	}

	if depFile != "" {
		err := deptools.WriteDepFile(depFile, outFile, deps)
		if err != nil {
			fatalf("error writing depfile: %s", err)
		}
	}

	if c, ok := config.(ConfigRemoveAbandonedFilesUnder); ok {
		under := c.RemoveAbandonedFilesUnder()
		err := removeAbandonedFilesUnder(ctx, bootstrapConfig, SrcDir, under)
		if err != nil {
			fatalf("error removing abandoned files: %s", err)
		}
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			fatalf("error opening memprofile: %s", err)
		}
		defer f.Close()
		pprof.WriteHeapProfile(f)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("\n")
	os.Exit(1)
}

func fatalErrors(errs []error) {
	red := "\x1b[31m"
	unred := "\x1b[0m"

	for _, err := range errs {
		switch err := err.(type) {
		case *blueprint.BlueprintError,
			*blueprint.ModuleError,
			*blueprint.PropertyError:
			fmt.Printf("%serror:%s %s\n", red, unred, err.Error())
		default:
			fmt.Printf("%sinternal error:%s %s\n", red, unred, err)
		}
	}
	os.Exit(1)
}
