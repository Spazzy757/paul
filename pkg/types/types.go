package types

import (
	"log"

	"gopkg.in/yaml.v2"
)

//PaulConfig defines the struct for type
type PaulConfig struct {
	Maintainers  []string     `yaml:"maintainers"`
	PullRequests PullRequests `yaml:"pull_requests"`
	Labels       bool         `yaml:"labels"`
}

//PullRequests struct
type PullRequests struct {
	OpenMessage string `yaml:"open_message"`
	CatsEnabled bool   `yaml:"cats_enabled"`
	DogsEnabled bool   `yaml:"dogs_enabled"`
}

//LoadConfig loads the config for the type PaulConfig
func (pc *PaulConfig) LoadConfig(config []byte) {
	err := yaml.Unmarshal(config, pc)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
