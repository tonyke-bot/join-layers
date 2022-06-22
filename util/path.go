package util

import "path/filepath"

func WithoutExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}
