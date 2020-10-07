package github

import (
	"context"
	"fmt"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"log"
)

const configFile = "PAUL.yaml"

func getPaulConfig(
	owner, repo *string,
	client *github.Client,
	contentUrl string,
	ctx context.Context,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig

	response, err := client.Repositories.DownloadContents(
		ctx,
		*owner,
		*repo,
		configFile,
		&github.RepositoryContentGetOptions{
			Ref: "master",
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

func RunPullRequestChecks(event *github.PullRequestEvent) {
	client, ctx := getClient(*event.Installation.ID)
	cfg, err := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		client,
		event.Repo.GetContentsURL(),
		ctx,
	)
	if err != nil {
		log.Fatalf("An error occurred fetching config %v", err)
	}
	if cfg.PullRequests.OpenMessage != "" && *event.Action == "opened" {
		pr := &pullRequestClient{ctx: ctx, client: client.PullRequests}
		comment(event.GetPullRequest(), pr, cfg.PullRequests.OpenMessage)
	}
}

type pullRequest interface {
	CreateReview(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		review *github.PullRequestReviewRequest,
	) (*github.PullRequestReview, *github.Response, error)
}

type pullRequestClient struct {
	ctx    context.Context
	client pullRequest
}

func comment(pr *github.PullRequest, client *pullRequestClient, message string) error {
	pullRequestReviewRequest := &github.PullRequestReviewRequest{
		Body:  &message,
		Event: github.String("COMMENT"),
	}

	_, _, err := client.client.CreateReview(
		client.ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		pullRequestReviewRequest,
	)
	if err != nil {
		return err
	}
	return nil
}
