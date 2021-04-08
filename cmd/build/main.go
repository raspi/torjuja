package main

import (
	"fmt"
	"github.com/raspi/torjuja/pkg/httpapi/frontend"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
	"os"
	"path"
)

func main() {
	converter := typescriptify.New()
	converter.CreateFromMethod = false
	converter.DontExport = false

	converter.Add(frontend.AllowDTO{})
	converter.Add(frontend.ResponseDTO{})

	err := converter.ConvertToFile(path.Join(`frontend`, `src`, `dto.ts`))
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}
}
