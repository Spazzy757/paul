package github

import (
	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"log"
)

//PullRequestHandler handler for the pull request event
func PullRequestHandler(
	event *github.PullRequestEvent,
	client *paulclient.GithubClient,
) error {
	// Get Paul Config
	cfg, configErr := config.GetPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		event.Repo.GetDefaultBranch(),
		client,
	)
	if configErr != nil {
		return configErr
	}
	var err error
	if firstPRCheck(cfg.PullRequests.OpenMessage, *event.Action) {
		err = reviewComment(
			event.GetPullRequest(),
			client,
			cfg.PullRequests.OpenMessage,
		)
	}
	if branchDestroyerCheck(
		&cfg.BranchDestroyer,
		*event.Action,
		event.Repo.GetDefaultBranch(),
		event.PullRequest.Head.GetRef(),
	) {
		log.Println("HERE")
		err = branchDestroyer(
			event.GetPullRequest(),
			client,
			event.PullRequest.Head.GetRef(),
		)
	}
	return err

}

//firstPRCheck checks if a PR has just been opened and
func firstPRCheck(message, action string) bool {
	return message != "" && action == "opened"
}

//branchDestroyerCheck checks if branch can be destroyed
func branchDestroyerCheck(
	cfg *types.BranchDestroyer,
	action, defaultBranch, destroyBranch string,
) bool {
	return cfg.Enabled &&
		action == "completed" &&
		destroyBranch != defaultBranch &&
		!checkStringInList(cfg.ProtectedBranches, destroyBranch)
}

//reviewComment sends a review comment to a Pull Request
func reviewComment(
	pr *github.PullRequest,
	client *paulclient.GithubClient,
	message string,
) error {
	pullRequestReviewRequest := &github.PullRequestReviewRequest{
		Body:  &message,
		Event: github.String("COMMENT"),
	}
	_, _, err := client.PullRequestService.CreateReview(
		client.Ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		pullRequestReviewRequest,
	)
	return err
}

//branchDestroyer will delete a branch
func branchDestroyer(
	pr *github.PullRequest,
	client *paulclient.GithubClient,
	branch string,
) error {
	_, err := client.GitService.DeleteRef(
		client.Ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		branch,
	)
	return err
}
