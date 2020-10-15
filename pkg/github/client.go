package github

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"

	"golang.org/x/oauth2"
)

const configFile = "PAUL.yaml"

/*
Interfaces to keep track of git client functions used
and to make testing easier
*/

//repoService handles functions specific to the repo
type repoService interface {
	DownloadContents(
		ctx context.Context,
		owner, repo, filepath string,
		opt *github.RepositoryContentGetOptions,
	) (io.ReadCloser, error)
}

//pullRequestService handles functions specific to Pull Requests
type pullRequestService interface {
	CreateReview(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		review *github.PullRequestReviewRequest,
	) (*github.PullRequestReview, *github.Response, error)
}

//gitService handles functions associated with git functionality
type gitService interface {
	DeleteRef(
		ctx context.Context,
		owner, repo, ref string,
	) (*github.Response, error)
}

/*
issueService handles functionality specific to issues
This includes Pull requests as some of the
underlying functionality is the same
*/
type issueService interface {
	CreateComment(
		ctx context.Context,
		owner, repo string,
		number int,
		comment *github.IssueComment,
	) (*github.IssueComment, *github.Response, error)
	AddLabelsToIssue(
		ctx context.Context,
		owner, repo string,
		number int,
		labels []string,
	) ([]*github.Label, *github.Response, error)
	RemoveLabelForIssue(
		ctx context.Context,
		owner, repo string,
		number int,
		label string,
	) (*github.Response, error)
}

/*
Structs to handle the different clients
*/
//repoClient handler for repoServices
type repoClient struct {
	ctx         context.Context
	repoService repoService
}

//gitClient handler for gitServices
type gitClient struct {
	ctx        context.Context
	gitService gitService
}

//pullRequestClient handler for pullRequestService
type pullRequestClient struct {
	ctx                context.Context
	pullRequestService pullRequestService
}

//issueClient handler for the issueService
type issueClient struct {
	ctx          context.Context
	issueService issueService
}

//getClient returns an authorized Github Client
func getClient(installationId int64) (*github.Client, context.Context, error) {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		return &github.Client{}, ctx, err
	}
	token, tokenErr := helpers.GetAccessToken(cfg, installationId)
	if tokenErr != nil {
		return &github.Client{}, ctx, tokenErr
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, ctx, nil
}

// TODO: Move This Logic into configs
//getPaulConig returns configuration for paul
func getPaulConfig(
	owner, repo *string,
	contentUrl, defaultBranch string,
	client *repoClient,
) (types.PaulConfig, error) {
	var paulCfg types.PaulConfig

	response, err := client.repoService.DownloadContents(
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
	err = paulCfg.LoadConfig(bytesConfig)
	return paulCfg, err
}
