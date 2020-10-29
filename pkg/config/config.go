package config

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
)

const configFile = "PAUL.yaml"

//GetPaulConfig returns configuration for paul
func GetPaulConfig(
	ctx context.Context,
	owner, repo, defaultBranch string,
	client *github.Client,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig

	response, err := client.Repositories.DownloadContents(
		ctx,
		owner,
		repo,
		configFile,
		&github.RepositoryContentGetOptions{
			Ref: defaultBranch,
		},
	)
	if err != nil {
		return paulCfg, fmt.Errorf("unable to download config file: %s", err)
	}
	defer response.Close()

	bytesConfig, err := ioutil.ReadAll(response)
	if err != nil {
		return paulCfg, fmt.Errorf("unable to read github's response: %s", err)
	}
	err = paulCfg.LoadConfig(bytesConfig)
	return paulCfg, err
}
