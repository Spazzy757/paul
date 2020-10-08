package types

import (
	"gopkg.in/yaml.v2"
	"log"
)

type PaulConfig struct {
	Maintainers  []string     `yaml:"maintainers"`
	PullRequests PullRequests `yaml:"pull_requests"`
}

type PullRequests struct {
	OpenMessage string `yaml:"open_message"`
	CatsEnabled bool   `yaml:"cats_enabled"`
}

func (pc *PaulConfig) LoadConfig(config []byte) {
	err := yaml.Unmarshal(config, pc)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
