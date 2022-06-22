package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"join-layers/util"
	"net/url"
	"os"
)

type Trait struct {
	Name        string  `yaml:"name"`
	DisplayName *string `yaml:"display_name,omitempty"`
}

type TraitSet struct {
	Name        *string `yaml:"name,omitempty"`
	Size        uint    `yaml:"size"`
	TraitsOrder []Trait `yaml:"traits"`
}

type Config struct {
	Name             string         `yaml:"name"`
	Description      string         `yaml:"description"`
	BaseURI          string         `yaml:"base_uri"`
	StartID          *uint          `yaml:"start_id,omitempty"`
	TraitSets        []TraitSet     `yaml:"trait_sets"`
	Width            uint           `yaml:"width"`
	Height           uint           `yaml:"height"`
	AdditionalData   map[string]any `yaml:"additional_data"`
	IsSolana         bool           `yaml:"is_solana"`
	CreatorAddress   *string        `yaml:"solana_creator_address"`
	CheckDuplication bool           `yaml:"check_duplication"`

	TypedBaseURI *url.URL `yaml:"-"`
}

func (c *Config) ActualStartID() uint {
	if c.StartID != nil {
		return *c.StartID
	}

	return 0
}

var ConfigTemplate = Config{
	Name:             "NFT Collection",
	Description:      "Your first ever NFT collection",
	BaseURI:          "ipfs://replace-with-your-ipfs-or-https-url/",
	StartID:          util.VarPtr[uint](0),
	CheckDuplication: true,
	TraitSets: []TraitSet{
		{
			Size: 100,
			TraitsOrder: []Trait{
				{Name: "Background"},
				{Name: "Eyeball"},
				{Name: "Eye color"},
				{Name: "Iris"},
				{Name: "Shine"},
				{Name: "Bottom lid"},
				{Name: "Top lid"},
			},
		},
	},
	Width:          2048,
	Height:         2048,
	AdditionalData: map[string]any{},
}

func Load(filePath string) *Config {
	f, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Fail to load config %v: %v", filePath, err)
		os.Exit(-1)
	}

	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		fmt.Printf("Fail to load config %v: %v", filePath, err)
		os.Exit(-1)
	}

	cfg.TypedBaseURI, err = url.Parse(cfg.BaseURI)
	if err != nil {
		fmt.Printf("Invalid baseURI %v\r\n", cfg.BaseURI)
		os.Exit(-1)
	}

	return &cfg
}
