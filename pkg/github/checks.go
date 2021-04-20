package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v34/github"
)

const (
	dco     = "DeveloperCertificateOfOrigin"
	success = "success"
	started = "in_progress"
	failed  = "action_required"
)

var isAnonymousSignature = regexp.MustCompile("Signed-off-by:(.*)noreply.github.com")

func getPullRequestCommits(
	ctx context.Context,
	event *github.PullRequestEvent,
	client *github.Client,
) ([]*github.RepositoryCommit, error) {
	listOpts := &github.ListOptions{
		Page: 0,
	}
	pr := event.PullRequest
	owner := pr.Base.Repo.Owner.GetLogin()
	repo := pr.Base.Repo.GetName()
	commits, resp, err := client.PullRequests.ListCommits(
		ctx,
		owner,
		repo,
		pr.GetNumber(),
		listOpts,
	)

	helpers.LogRateLimit(
		"ListCommits",
		resp.Rate.Limit,
		resp.Rate.Remaining,
	)

	if err != nil {
		return nil, err
	}

	resp.Body.Close()
	return commits, nil
}

func createDCOCheck(
	event *github.PullRequestEvent,
) github.CreateCheckRunOptions {
	now := github.Timestamp{Time: time.Now()}
	status := started
	text := "Checking Developer Certificate of Origin"
	title := "Developer Certificate of Origin"
	summary := "Checking Commits Are Signed"
	check := github.CreateCheckRunOptions{
		StartedAt: &now,
		Name:      dco,
		HeadSHA:   event.PullRequest.Head.GetSHA(),
		Status:    &status,
		Output: &github.CheckRunOutput{
			Text:    &text,
			Title:   &title,
			Summary: &summary,
		},
	}
	conclusion := success
	check.Conclusion = &conclusion
	check.CompletedAt = &now
	return check
}

func updateUnsuccessfulDCOCheck(
	check *github.CheckRun,
) github.UpdateCheckRunOptions {
	now := github.Timestamp{Time: time.Now()}
	text := `Thank you for your contribution, please make sure you have signed off all your commits`
	title := "Unsigned commits"
	summary := "One or more of the commits in this Pull Request are not signed-off."

	checkOpt := github.UpdateCheckRunOptions{
		Name: check.GetName(),
		Output: &github.CheckRunOutput{
			Text:    &text,
			Title:   &title,
			Summary: &summary,
		},
	}
	conclusion := failed
	checkOpt.Conclusion = &conclusion
	checkOpt.CompletedAt = &now
	return checkOpt
}

func updateSuccessfulDCOCheck(
	check *github.CheckRun,
) github.UpdateCheckRunOptions {
	now := github.Timestamp{Time: time.Now()}
	text := "Thank you for the contribution, everything looks fine."
	title := "Signed commits"
	summary := "All of your commits are signed"

	checkOpt := github.UpdateCheckRunOptions{
		Name: check.GetName(),
		Output: &github.CheckRunOutput{
			Text:    &text,
			Title:   &title,
			Summary: &summary,
		},
	}
	conclusion := success
	checkOpt.Conclusion = &conclusion
	checkOpt.CompletedAt = &now
	return checkOpt
}

func getExistingDCOCheck(
	ctx context.Context,
	event *github.PullRequestEvent,
	client *github.Client,
) (*github.ListCheckRunsResults, error) {
	pr := event.PullRequest
	checkName := dco
	checks, res, err := client.Checks.ListCheckRunsForRef(ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		pr.Head.GetSHA(),
		&github.ListCheckRunsOptions{CheckName: &checkName})

	if res.StatusCode != 200 {
		return nil, fmt.Errorf(
			"Error unexpected status code while retreiving existing checks %d",
			res.StatusCode,
		)
	}
	return checks, err
}

func createSuccessfulCheck(
	ctx context.Context,
	event *github.PullRequestEvent,
	client *github.Client,
) (*github.CheckRun, error) {
	checks, err := getExistingDCOCheck(ctx, event, client)
	if err != nil {
		return &github.CheckRun{}, err
	}
	if checks.GetTotal() > 1 {
		return &github.CheckRun{}, fmt.Errorf("Error unexpected count of existing DCO checks: %d", *checks.Total)
	}
	if checks.GetTotal() == 1 {
		return checks.CheckRuns[0], nil
	}
	pr := event.PullRequest
	newCheck := createDCOCheck(event)
	check, res, err := client.Checks.CreateCheckRun(
		ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		newCheck,
	)
	if err != nil {
		return check, err
	}
	if res.StatusCode != 201 {
		return check, fmt.Errorf("DCO check unexpected status code: %d", res.StatusCode)
	}
	return check, nil
}

func updateExistingDCOCheck(
	ctx context.Context,
	client *github.Client,
	event *github.PullRequestEvent,
	check *github.CheckRun,
	conclusion string,
) error {
	var checkOpts github.UpdateCheckRunOptions

	if conclusion == success {
		checkOpts = updateSuccessfulDCOCheck(check)
	} else if conclusion == failed {
		checkOpts = updateUnsuccessfulDCOCheck(check)
	}

	pr := event.PullRequest
	_, resp, err := client.Checks.UpdateCheckRun(
		ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		check.GetID(),
		checkOpts,
	)
	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"Error while updating the DCO check unexpected status code: %d",
			resp.StatusCode,
		)
	}
	if err != nil {
		return err
	}
	return nil
}

func hasUnsigned(commits []*github.RepositoryCommit) bool {
	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			return !isSigned(*commit.Commit.Message)
		}
	}
	return false
}

func isSigned(msg string) bool {
	return strings.Contains(msg, "Signed-off-by:")
}

func hasAnonymousSign(commits []*github.RepositoryCommit) bool {
	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			if isAnonymousSign(*commit.Commit.Message) {
				return true
			}
		}
	}
	return false
}

func isAnonymousSign(msg string) bool {
	return isAnonymousSignature.Match([]byte(msg))
}
