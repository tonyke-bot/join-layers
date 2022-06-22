package initialize

import (
	"fmt"
	"github.com/akamensky/argparse"
	"gopkg.in/yaml.v3"
	"join-layers/config"
	"join-layers/util"
	"os"
)

var initConfigFile *string

func Setup(parser *argparse.Parser) *argparse.Command {
	initCmd := parser.NewCommand("init", "initialize the config file")
	initConfigFile = initCmd.String("c", "config", &argparse.Options{
		Required: false,
		Validate: util.ValidateStringArgs,
		Help:     "path to the config file. default: config.yaml",
		Default:  "config.yaml",
	})

	return initCmd
}

func Exec() {
	f, err := os.Create(*initConfigFile)
	if err != nil {
		fmt.Printf("Fail to open file: %v\r\n", err)
		os.Exit(-1)
	}
	defer f.Close()

	err = yaml.NewEncoder(f).Encode(config.ConfigTemplate)
	if err != nil {
		fmt.Printf("Fail to encode yaml: %v\r\n", err)
		os.Exit(-1)
	}
}
