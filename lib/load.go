package lib

import (
	"encoding/json"
	"os"
)

func Load(file string) (*Config, error) {
	cfg := &Config{}
	fd, err := os.Open(file)
	if err != nil {
		return cfg, err
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)

	err = dec.Decode(cfg)
	return cfg, err
}
