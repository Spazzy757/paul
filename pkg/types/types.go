package types

import (
	"gopkg.in/yaml.v2"
)

//PaulConfig defines the struct for type
type PaulConfig struct {
	Maintainers           []string              `yaml:"maintainers,omitempty"`
	PullRequests          PullRequests          `yaml:"pull_requests,omitempty"`
	Labels                bool                  `yaml:"labels,omitempty"`
	BranchDestroyer       BranchDestroyer       `yaml:"branch_destroyer,omitempty"`
	EmptyDescriptionCheck EmptyDescriptionCheck `yaml:"empty_description_check,omitempty"`
}

//EmptyDescriptionCheck config for empty PR checks
type EmptyDescriptionCheck struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Enforced bool   `yaml:"enforced,omitempty"`
	Message  string `yaml:"message,omitempty"`
}

//PullRequests struct
type PullRequests struct {
	OpenMessage         string            `yaml:"open_message,omitempty"`
	AllowApproval       bool              `yaml:"allow_approval,omitempty"`
	Assign              bool              `yaml:"assign,omitempty"`
	StaleTime           int               `yaml:"stale_time,omitempty"`
	CatsEnabled         bool              `yaml:"cats_enabled,omitempty"`
	DogsEnabled         bool              `yaml:"dogs_enabled,omitempty"`
	GiphyEnabled        bool              `yaml:"giphy_enabled,omitempty"`
	AutomatedMerge      bool              `yaml:"automated_merge"`
	LimitPullRequests   LimitPullRequests `yaml:"limit_pull_requests,omitempty"`
	DCOCheck            bool              `yaml:"dco_check,omitempty"`
	VerifiedCommitCheck bool              `yaml:"verified_commit_check,omitempty"`
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
