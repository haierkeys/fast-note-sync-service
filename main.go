package main

import (
	"embed"

	"github.com/haierkeys/fast-note-sync-service/cmd"
)

//go:embed frontend
var efs embed.FS

//go:embed config/config.yaml
var c string

func main() {
	cmd.Execute(efs, c)
}
