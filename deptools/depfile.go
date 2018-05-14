package deptools

import (
	"fmt"
	"os"
	"strings"
)

var (
	pathEscaper = strings.NewReplacer(
		`\`, `\\`,
		` `, `\ `,
		`#`, `\#`,
		`*`, `\*`,
		`[`, `\[`,
		`|`, `\|`)
)

// WriteDepFile creates a new gcc-style depfile and
// populates it with content indicating that target
// depends on deps.
func WriteDepFile(filename, target string, deps []string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var escapedDeps []string

	for _, dep := range deps {
		escapedDeps = append(escapedDeps, pathEscaper.Replace(dep))
	}

	_, err = fmt.Fprintf(f, "%s: \\\n %s\n", target,
		strings.Join(escapedDeps, " \\\n "))
	if err != nil {
		return err
	}

	return nil
}
