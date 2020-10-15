package types

import (
	"gopkg.in/yaml.v2"
)

//PaulConfig defines the struct for type
type PaulConfig struct {
	Maintainers     []string        `yaml:"maintainers"`
	PullRequests    PullRequests    `yaml:"pull_requests"`
	Labels          bool            `yaml:"labels"`
	BranchDestroyer BranchDestroyer `yaml:"branch_destroyer"`
}

//PullRequests struct
type PullRequests struct {
	OpenMessage string `yaml:"open_message"`
	CatsEnabled bool   `yaml:"cats_enabled"`
	DogsEnabled bool   `yaml:"dogs_enabled"`
}

//PullRequests struct
type BranchDestroyer struct {
	Enabled           bool     `yaml:"enabled"`
	ProtectedBranches []string `yaml:"protected_branches"`
}

//LoadConfig loads the config for the type PaulConfig
func (pc *PaulConfig) LoadConfig(config []byte) error {
	err := yaml.Unmarshal(config, pc)
	return err
}
