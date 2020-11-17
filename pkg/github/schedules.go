package github

import (
	"context"
	"time"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	log "github.com/sirupsen/logrus"
)

const (
	staleLabel = "stale"
	mergeLabel = "merge"
)

// ScehduledPullRequests helper struct to limit calls to repos
type ScehduledJobInformation struct {
	Cfg          types.PaulConfig
	Repo         *github.Repository
	PullRequests []*github.PullRequest
}

// PullRequestsScheduledJobs will go through an installations repos
// and run jobs against all Pull Requests where Paul is installed
func PullRequestsScheduledJobs(
	ctx context.Context,
	client *github.Client,
) {
	scheduledJobsInformationList, err := getScehduledJobInformationList(ctx, client)
	if handleError(err) {
		return
	}
	// Check if Pull Requests Should Be Marked as Stale
	markPullRequestsStale(ctx, client, scheduledJobsInformationList)
	// Merges Pull Requests that are viable
	mergePendingPullRequests(ctx, client, scheduledJobsInformationList)
}

func markPullRequestsStale(
	ctx context.Context,
	client *github.Client,
	informationList []*ScehduledJobInformation,
) {
	var stalePullRequests []*github.PullRequest
	for _, scheduledJobsInformation := range informationList {
		cfg := scheduledJobsInformation.Cfg
		if cfg.PullRequests.StaleTime != 0 {
			stalePullRequests = append(
				stalePullRequests,
				checkTimeStamps(cfg, scheduledJobsInformation.PullRequests)...,
			)
		}
	}
	for _, pullRequest := range stalePullRequests {
		err := markPullRequestStale(ctx, client, pullRequest)
		if handleError(err) {
			continue
		}
	}
}

func mergePendingPullRequests(
	ctx context.Context,
	client *github.Client,
	informationList []*ScehduledJobInformation,
) {
	var labeledPullRequests []*github.PullRequest
	for _, scheduledJobsInformation := range informationList {
		cfg := scheduledJobsInformation.Cfg
		if cfg.PullRequests.AutomatedMerge {
			labeledPullRequests = append(
				labeledPullRequests,
				checkLabels(mergeLabel, scheduledJobsInformation.PullRequests)...,
			)
		}
	}
	for _, pullRequest := range labeledPullRequests {
		if pullRequest.GetMergeable() {
			err := mergePullRequest(ctx, client, pullRequest)
			if handleError(err) {
				continue
			}
		}
	}
}

// Returns for scheduled jobs in a list
// this reduces the amount of calls needed to each repo
func getScehduledJobInformationList(
	ctx context.Context,
	client *github.Client,
) ([]*ScehduledJobInformation, error) {
	// Returns all the repos for this installation
	// limited to 50
	var scheduledJobInformations []*ScehduledJobInformation
	repos, _, err := client.Apps.ListRepos(
		ctx,
		// TODO: Remove Limit on Repos
		&github.ListOptions{PerPage: 50},
	)
	for _, repo := range repos {
		// we ignore th error as it will return an empty paul config
		// which will do nothing
		cfg, err := config.GetPaulConfig(
			ctx,
			repo.Owner.GetLogin(),
			repo.GetName(),
			repo.GetDefaultBranch(),
			client,
		)
		// If error no fetching config carry on
		// otherwise their is either no config or a connection problem
		if err == nil {
			pullRequests, pullRequestErr := listPullRequests(ctx, client, repo)
			if handleError(pullRequestErr) {
				continue
			}
			scheduledJobInformations = append(
				scheduledJobInformations,
				&ScehduledJobInformation{
					Cfg:          cfg,
					Repo:         repo,
					PullRequests: pullRequests,
				},
			)
		}
	}
	return scheduledJobInformations, err
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
	// We use the issues client as Pull Requests are treated as Issues
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

func checkLabels(
	label string,
	prs []*github.PullRequest,
) []*github.PullRequest {
	var labeledPullRequests []*github.PullRequest
	for _, pr := range prs {
		for _, pullRequestLabel := range pr.Labels {
			if pullRequestLabel.GetName() == label {
				labeledPullRequests = append(labeledPullRequests, pr)
			}
		}
	}
	return labeledPullRequests
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
