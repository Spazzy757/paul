package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v36/github"
)

const configFile = "PAUL.yaml"

//GetPaulConfig returns configuration for paul
func GetPaulConfig(
	ctx context.Context,
	owner, repo, defaultBranch string,
	client *github.Client,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig
	reader, resp, err := client.Repositories.DownloadContents(
		ctx,
		owner,
		repo,
		filepath.Join(".github", configFile),
		&github.RepositoryContentGetOptions{
			Ref: defaultBranch,
		},
	)
	// If 404 check in the root directory
	if resp.StatusCode == 404 {
		reader, resp, err = client.Repositories.DownloadContents(
			ctx,
			owner,
			repo,
			configFile,
			&github.RepositoryContentGetOptions{
				Ref: defaultBranch,
			},
		)
		// If still not found then return empty config
		// but not error
		if resp.StatusCode == 404 {
			return paulCfg, nil
		}
	}
	// Any error from downloading return empty config
	// This means the file cant be found or paul does not have access
	if err != nil {
		return paulCfg, nil

	}
	defer reader.Close()

	bytesConfig, err := ioutil.ReadAll(reader)
	if err != nil {
		return paulCfg, fmt.Errorf("unable to read github's response: %s", err)
	}
	err = paulCfg.LoadConfig(bytesConfig)
	return paulCfg, err
}
