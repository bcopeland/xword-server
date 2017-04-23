package conf

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	DBUri string `toml:"dburi"`
}

func LoadConfig(filename string) (config Config, err error) {
	_, err = toml.DecodeFile(filename, &config)
	return config, err
}
