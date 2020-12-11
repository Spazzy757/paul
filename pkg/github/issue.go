package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
)

/*
IssueCommentHandler takes an incoming event of type IssueCommentEvent and
runs logic against it
*/
func IssueCommentHandler(
	ctx context.Context,
	event *github.IssueCommentEvent,
	client *github.Client,
) error {
	// load Paul Config from repo
	cfg, configErr := config.GetPaulConfig(
		ctx,
		event.Repo.Owner.GetLogin(),
		event.Repo.GetName(),
		event.Repo.GetDefaultBranch(),
		client,
	)
	if configErr != nil {
		return configErr
	}

	var err error
	// Check comments for any commands
	if *event.Action == "created" {
		// Get Comment
		comment := event.GetComment()
		// Get Which Command is run
		// Throw away args as they are not used currently
		cmd, args := getCommand(*comment.Body)
		// Switch statement to handle different commands
		switch {
		// Case of /cat command
		case cmd == "cat" && cfg.PullRequests.CatsEnabled:
			// Get the Cat Client
			animalClient := animals.NewCatClient()
			err = catsHandler(ctx, event, client, animalClient)
		// Case of /dog command
		case cmd == "dog" && cfg.PullRequests.DogsEnabled:
			// Get the Dog Client
			animalClient := animals.NewDogClient()
			err = dogsHandler(ctx, event, client, animalClient)
		// Case /label command
		case cmd == "label":
			// handle the labels
			// only a single label can be added at a time
			// i.e "good first issue"
			labels := []string{strings.Join(args, " ")}
			err = labelHandler(ctx, &cfg, event, client, labels)
		// Case /remove-label command
		case cmd == "remove-label":
			err = removeLabelHandler(ctx, &cfg, event, client, args)
		// Case /approve command
		case cmd == "approve":
			err = approveHandler(ctx, &cfg, event, client)
		// Case /merge command
		case cmd == "merge":
			err = mergeHandler(ctx, &cfg, event, client)
		default:
			break
		}
	}
	return err
}

// getCommand strips out the command and any args that are given
func getCommand(comment string) (string, []string) {
	var args []string
	if !strings.HasPrefix(comment, "/") {
		return "", args
	}
	commands := strings.Split(comment[1:], " ")
	return commands[0], commands[1:]
}

// handler for the /merge command
func mergeHandler(
	ctx context.Context,
	cfg *types.PaulConfig,
	event *github.IssueCommentEvent,
	client *github.Client,
) error {
	if event.Issue.IsPullRequest() &&
		checkStringInList(cfg.Maintainers, event.Sender.GetLogin()) {
		pr, _, err := client.PullRequests.Get(
			ctx,
			event.Repo.Owner.GetLogin(),
			event.Repo.GetName(),
			event.Issue.GetNumber(),
		)
		if err != nil {
			return err
		}
		if pr.GetMergeable() {
			err = mergePullRequest(ctx, client, pr)
		} else {
			message := "This Pull Request Can not be merge currently"
			err = createIssueComment(ctx, event, client, message)
		}
		return err
	}
	return nil
}

// handleCats is the handler for the /cat command
func catsHandler(
	ctx context.Context,
	is *github.IssueCommentEvent,
	client *github.Client,
	catClient *animals.Client,
) error {
	cat, err := catClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("My Most Trusted Minion\n\n ![my favorite minion](%v)", cat.Url)
	err = createIssueComment(ctx, is, client, message)
	return err
}

// handleDogs is the handler for the /dog command
func dogsHandler(
	ctx context.Context,
	is *github.IssueCommentEvent,
	client *github.Client,
	dogClient *animals.Client,
) error {
	dog, err := dogClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Despite how it looks it is well trained\n\n ![loyal soldier](%v)", dog.Url)
	err = createIssueComment(ctx, is, client, message)
	return err
}

//labelHandler handles the /label command
func labelHandler(
	ctx context.Context,
	cfg *types.PaulConfig,
	event *github.IssueCommentEvent,
	client *github.Client,
	labels []string,
) error {
	var err error
	if cfg.Labels &&
		checkStringInList(cfg.Maintainers, event.Sender.GetLogin()) {
		_, _, err = client.Issues.AddLabelsToIssue(
			ctx,
			event.Repo.Owner.GetLogin(),
			event.Repo.GetName(),
			event.Issue.GetNumber(),
			labels,
		)
	}
	return err
}

//removeLabelHandler handles the /removelabel command
func removeLabelHandler(
	ctx context.Context,
	cfg *types.PaulConfig,
	event *github.IssueCommentEvent,
	client *github.Client,
	labels []string,
) error {
	var err error
	if cfg.Labels &&
		// handle the remove labels,
		// if more than one arg is passed through, don't do anything
		len(labels) == 1 &&
		checkStringInList(cfg.Maintainers, event.Sender.GetLogin()) {
		_, err = client.Issues.RemoveLabelForIssue(
			ctx,
			event.Repo.Owner.GetLogin(),
			event.Repo.GetName(),
			event.Issue.GetNumber(),
			labels[0],
		)
	}
	return err
}

//approveHandler approves Pull Requests
func approveHandler(
	ctx context.Context,
	cfg *types.PaulConfig,
	event *github.IssueCommentEvent,
	client *github.Client,
) error {
	var err error
	if cfg.PullRequests.AllowApproval &&
		checkStringInList(cfg.Maintainers, event.Sender.GetLogin()) &&
		event.Issue.IsPullRequest() {
		pullRequestReviewRequest := &github.PullRequestReviewRequest{
			Event: github.String("APPROVE"),
		}
		_, _, err = client.PullRequests.CreateReview(
			ctx,
			event.Repo.Owner.GetLogin(),
			event.Repo.GetName(),
			event.Issue.GetNumber(),
			pullRequestReviewRequest,
		)
	}
	return err
}

// createIssueComment sends a comment to an issue/pull request
func createIssueComment(
	ctx context.Context,
	is *github.IssueCommentEvent,
	client *github.Client,
	message string,
) error {
	comment := &github.IssueComment{Body: &message}
	_, _, err := client.Issues.CreateComment(
		ctx,
		*is.Repo.Owner.Login,
		is.Repo.GetName(),
		is.Issue.GetNumber(),
		comment,
	)
	return err
}

// checkStringInList checks if string is in a list of strings
func checkStringInList(stringList []string, query string) bool {
	for _, i := range stringList {
		if i == query {
			return true
		}
	}
	return false
}
