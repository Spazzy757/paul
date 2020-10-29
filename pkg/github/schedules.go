package github

import (
	"context"
	"time"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	log "github.com/sirupsen/logrus"
)

const staleLabel = "stale"

// MarkPullRequestsAsStale will go through an installations repos
// and mark Pull Requests as stale if updated at is past a certain
// amount of days
func MarkPullRequestsAsStale(
	ctx context.Context,
	client *github.Client,
) {
	// Returns all the repos for this installation
	// limited to 50
	repos, _, err := client.Apps.ListRepos(
		ctx,
		// TODO: Remove Limit on Repos
		&github.ListOptions{PerPage: 50},
	)
	if handleError(err) {
		return
	}
	var stalePullRequests []*github.PullRequest
	for _, repo := range repos {
		// we ignore th error as it will return an empty paul config
		// which will do nothing
		cfg, _ := config.GetPaulConfig(
			ctx,
			repo.Owner.GetLogin(),
			repo.GetName(),
			repo.GetDefaultBranch(),
			client,
		)
		if cfg.PullRequests.StaleTime != 0 {
			pullRequests, err := listPullRequests(ctx, client, repo)
			if handleError(err) {
				return
			}
			stalePullRequests = append(
				stalePullRequests,
				checkTimeStamps(cfg, pullRequests)...,
			)
		}
	}
	for _, pullRequest := range stalePullRequests {
		err = markPullRequestStale(ctx, client, pullRequest)
		if handleError(err) {
			return
		}
	}
}

func handleError(err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Scheduling error occurred")
		return true
	}
	return false
}

func markPullRequestStale(
	ctx context.Context,
	client *github.Client,
	pullRequest *github.PullRequest,
) error {
	// We use the issues client as it Pull Requests are treated as Issues
	_, _, err := client.Issues.AddLabelsToIssue(
		ctx,
		pullRequest.Base.Repo.Owner.GetLogin(),
		pullRequest.Base.Repo.GetName(),
		pullRequest.GetNumber(),
		[]string{staleLabel},
	)
	return err
}

func checkTimeStamps(
	cfg types.PaulConfig,
	prs []*github.PullRequest,
) []*github.PullRequest {
	now := time.Now()
	var stalePullRequests []*github.PullRequest
	for _, pr := range prs {
		days := int(now.Sub(pr.GetUpdatedAt()).Hours() / 24)
		if days > cfg.PullRequests.StaleTime {
			stalePullRequests = append(stalePullRequests, pr)
		}
	}
	return stalePullRequests
}

func listPullRequests(
	ctx context.Context,
	client *github.Client,
	repo *github.Repository,
) ([]*github.PullRequest, error) {
	prs, _, err := client.PullRequests.List(
		ctx,
		repo.Owner.GetLogin(),
		repo.GetName(),
		&github.PullRequestListOptions{},
	)
	return prs, err
}
