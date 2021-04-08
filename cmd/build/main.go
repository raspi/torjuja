package main

import (
	"github.com/raspi/torjuja/pkg/httpapi/frontend"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	converter := typescriptify.New()
	converter.CreateFromMethod = false
	converter.DontExport = false

	converter.Add(frontend.AllowDTO{})

	err := converter.ConvertToFile("frontend/src/dto.ts")
	if err != nil {
		panic(err.Error())
	}
}
