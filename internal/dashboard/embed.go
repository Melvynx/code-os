package dashboard

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist/*
var embeddedAssets embed.FS

func Assets() (fs.FS, error) {
	assets, err := fs.Sub(embeddedAssets, "dist")
	if err != nil {
		return nil, fmt.Errorf("load dashboard build: %w", err)
	}
	return assets, nil
}
