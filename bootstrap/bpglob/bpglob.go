package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/blueprint/pathtools"
)

var (
	out = flag.String("o", "", "file to write list of files that match glob")

	excludes multiArg
)

func init() {
	flag.Var(&excludes, "e", "pattern to exclude from results")
}

type multiArg []string

func (m *multiArg) String() string {
	return `""`
}

func (m *multiArg) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func (m *multiArg) Get() interface{} {
	return m
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: bpglob -o out glob")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	if *out == "" {
		fmt.Fprintln(os.Stderr, "error: -o is required")
		usage()
	}

	if flag.NArg() != 1 {
		usage()
	}

	_, err := pathtools.GlobWithDepFile(flag.Arg(0), *out, *out+".d", excludes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}
