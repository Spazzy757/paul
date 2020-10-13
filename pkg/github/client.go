package github

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"

	"golang.org/x/oauth2"
)

const configFile = "PAUL.yaml"

func getClient(installationId int64) (*github.Client, context.Context) {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Can't load config: %v", err)
	}
	token, tokenErr := helpers.GetAccessToken(cfg, installationId)
	if tokenErr != nil {
		log.Fatalf("Can't load config: %v", err)
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, ctx
}

// TODO: Move This Logic into configs

// interface to make testing logic easier
type repo interface {
	DownloadContents(
		ctx context.Context,
		owner, repo, filepath string,
		opt *github.RepositoryContentGetOptions,
	) (io.ReadCloser, error)
}

// struct to make testing logic easier
type repoClient struct {
	ctx    context.Context
	client repo
}

func getPaulConfig(
	owner, repo *string,
	contentUrl, defaultBranch string,
	client *repoClient,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig

	response, err := client.client.DownloadContents(
		client.ctx,
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
	paulCfg.LoadConfig(bytesConfig)
	return paulCfg, nil
}
