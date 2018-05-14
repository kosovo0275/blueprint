package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var (
	output = flag.String("o", "", "output filename")
	pkg    = flag.String("p", "main", "package name")
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "error: must pass at least one input")
		os.Exit(1)
	}

	buf := &bytes.Buffer{}

	err := pluginTmpl.Execute(buf, struct {
		Package string
		Plugins []string
	}{
		filepath.Base(*pkg),
		flag.Args(),
	})
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*output, buf.Bytes(), 0666)
	if err != nil {
		panic(err)
	}
}

var pluginTmpl = template.Must(template.New("pluginloader").Parse(`
package {{.Package}}

import (
{{range .Plugins}}
	_ "{{.}}"
{{end}}
)
`))
