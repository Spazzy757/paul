package config

import (
	"fmt"
	"io/ioutil"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
)

const configFile = "PAUL.yaml"

// TODO: Move This Logic into configs
//getPaulConig returns configuration for paul
func GetPaulConfig(
	owner, repo *string,
	contentUrl, defaultBranch string,
	client *paulclient.GithubClient,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig

	response, err := client.RepoService.DownloadContents(
		client.Ctx,
		*owner,
		*repo,
		configFile,
		&github.RepositoryContentGetOptions{
			Ref: "main",
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
