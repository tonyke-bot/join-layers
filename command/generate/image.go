package generate

import (
	"image"
)

type Image struct {
	Name      string      `yaml:"name"`
	LayerName string      `yaml:"layer_name"`
	FullPath  string      `yaml:"full_path"`
	Rarity    float64     `yaml:"rarity"`
	Hash      []byte      `yaml:"-"`
	Obj       image.Image `yaml:"-"`
}
