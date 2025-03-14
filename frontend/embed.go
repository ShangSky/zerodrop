package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

var StaticFS fs.FS

func init() {
	f, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	StaticFS = f
}
