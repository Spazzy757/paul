package types

import (
	"gopkg.in/yaml.v2"
)

//PaulConfig defines the struct for type
type PaulConfig struct {
	Maintainers     []string        `yaml:"maintainers,omitempty"`
	PullRequests    PullRequests    `yaml:"pull_requests,omitempty"`
	Labels          bool            `yaml:"labels,omitempty"`
	BranchDestroyer BranchDestroyer `yaml:"branch_destroyer,omitempty"`
}

//PullRequests struct
type PullRequests struct {
	OpenMessage       string            `yaml:"open_message,omitempty"`
	AllowApproval     bool              `yaml:"allow_approval,omitempty"`
	CatsEnabled       bool              `yaml:"cats_enabled,omitempty"`
	DogsEnabled       bool              `yaml:"dogs_enabled,omitempty"`
	LimitPullRequests LimitPullRequests `yaml:"limit_pull_requests,omitempty"`
}

//LimitPullRequests struct
type LimitPullRequests struct {
	MaxNumber int `yaml:"max_number,omitempty"`
}

//BranchDestroyer struct
type BranchDestroyer struct {
	Enabled           bool     `yaml:"enabled,omitempty"`
	ProtectedBranches []string `yaml:"protected_branches,omitempty"`
}

//LoadConfig loads the config for the type PaulConfig
func (pc *PaulConfig) LoadConfig(config []byte) error {
	err := yaml.Unmarshal(config, pc)
	return err
}
