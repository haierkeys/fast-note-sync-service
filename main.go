package main

import (
	"embed"

	"github.com/haierkeys/fast-note-sync-service/cmd"
)

//go:embed frontend docs
var efs embed.FS

//go:embed config/config.yaml
var c string

// @title Fast Note Sync Service HTTP API
// @version 1.0
// @description This is the Fast Note Sync Service HTTP API.

// @contact.name Haierkeys
// @contact.url https://github.com/haierkeys
// @contact.email haierkeys@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9000
// @BasePath /

// @securityDefinitions.apikey UserAuthToken
// @in header
// @name token

// @securityDefinitions.apikey ShareAuthToken
// @in header
// @name Share-Token
func main() {
	cmd.Execute(efs, c)
}
