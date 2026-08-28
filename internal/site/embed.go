package site

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist/* dist/assets/* dist/docs/* dist/skills-sync/* dist/skills-sync/local/* dist/skills-sync/vps/*
var embeddedAssets embed.FS

func Assets() (fs.FS, error) {
	assets, err := fs.Sub(embeddedAssets, "dist")
	if err != nil {
		return nil, fmt.Errorf("load public site build: %w", err)
	}
	return assets, nil
}
