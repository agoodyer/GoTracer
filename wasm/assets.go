//go:build js && wasm

package main

import (
	"embed"
)

//go:embed assets/earthmap.jpg assets/moon.jpg
var embeddedAssets embed.FS

// GetEmbeddedAsset retrieves embedded asset data by name
func GetEmbeddedAsset(name string) ([]byte, error) {
	return embeddedAssets.ReadFile("assets/" + name)
}
