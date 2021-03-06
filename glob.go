package blueprint

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

type GlobPath struct {
	Pattern  string
	Excludes []string
	Files    []string
	Deps     []string
	Name     string
}

func verifyGlob(fileName, pattern string, excludes []string, g GlobPath) {
	if pattern != g.Pattern {
		panic(fmt.Errorf("Mismatched patterns %q and %q for glob file %q", pattern, g.Pattern, fileName))
	}
	if len(excludes) != len(g.Excludes) {
		panic(fmt.Errorf("Mismatched excludes %v and %v for glob file %q", excludes, g.Excludes, fileName))
	}

	for i := range excludes {
		if g.Excludes[i] != excludes[i] {
			panic(fmt.Errorf("Mismatched excludes %v and %v for glob file %q", excludes, g.Excludes, fileName))
		}
	}
}

func (c *Context) glob(pattern string, excludes []string) ([]string, error) {
	fileName := globToFileName(pattern, excludes)

	// Try to get existing glob from the stored results
	c.globLock.Lock()
	g, exists := c.globs[fileName]
	c.globLock.Unlock()

	if exists {
		// Glob has already been done, double check it is identical
		verifyGlob(fileName, pattern, excludes, g)
		return g.Files, nil
	}

	// Get a globbed file list
	files, deps, err := c.fs.Glob(pattern, excludes)
	if err != nil {
		return nil, err
	}

	// Store the results
	c.globLock.Lock()
	if g, exists = c.globs[fileName]; !exists {
		c.globs[fileName] = GlobPath{pattern, excludes, files, deps, fileName}
	}
	c.globLock.Unlock()

	// Getting the list raced with another goroutine, throw away the results and use theirs
	if exists {
		verifyGlob(fileName, pattern, excludes, g)
		return g.Files, nil
	}

	return files, nil
}

func (c *Context) Globs() []GlobPath {
	fileNames := make([]string, 0, len(c.globs))
	for k := range c.globs {
		fileNames = append(fileNames, k)
	}
	sort.Strings(fileNames)

	globs := make([]GlobPath, len(fileNames))
	for i, fileName := range fileNames {
		globs[i] = c.globs[fileName]
	}

	return globs
}

func globToString(pattern string) string {
	ret := ""
	for _, c := range pattern {
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '_', c == '-', c == '/':
			ret += string(c)
		default:
			ret += "_"
		}
	}

	return ret
}

func globToFileName(pattern string, excludes []string) string {
	name := globToString(pattern)
	excludeName := ""
	for _, e := range excludes {
		excludeName += "__" + globToString(e)
	}

	// Prevent file names from reaching ninja's path component limit
	if strings.Count(name, "/")+strings.Count(excludeName, "/") > 30 {
		excludeName = fmt.Sprintf("___%x", md5.Sum([]byte(excludeName)))
	}

	return name + excludeName + ".glob"
}
