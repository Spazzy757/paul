package github

import (
	"fmt"
	"strings"

	"github.com/Spazzy757/paul/pkg/animals"
	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/google/go-github/v32/github"
)

/*
IssueCommentHandler takes an incoming event of type IssueCommentEvent and
runs logic against it
*/
func IssueCommentHandler(
	event *github.IssueCommentEvent,
	client *paulclient.GithubClient,
) error {
	// load Paul Config from repo
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
			err = catsHandler(event, client, animalClient)
		// Case of /dog command
		case cmd == "dog" && cfg.PullRequests.DogsEnabled:
			// Get the Dog Client
			animalClient := animals.NewDogClient()
			err = dogsHandler(event, client, animalClient)
		// Case /label command
		case cmd == "label" &&
			cfg.Labels &&
			checkStringInList(cfg.Maintainers, *event.Sender.Login):
			// handle the labels
			err = labelHandler(event, client, args)
		// Case /remove-label command
		case cmd == "remove-label" &&
			cfg.Labels &&
			checkStringInList(cfg.Maintainers, *event.Sender.Login):
			// handle the remove labels,
			// if more than one arg is passed through, don't do anything
			if len(args) == 1 {
				err = removeLabelHandler(event, client, args[0])
			}
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

// handleCats is the handler for the /cat command
func catsHandler(
	is *github.IssueCommentEvent,
	client *paulclient.GithubClient,
	catClient *animals.Client,
) error {
	cat, err := catClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("My Most Trusted Minion\n\n ![my favorite minion](%v)", cat.Url)
	err = createIssueComment(is, client, message)
	return err
}

// handleDogs is the handler for the /dog command
func dogsHandler(
	is *github.IssueCommentEvent,
	client *paulclient.GithubClient,
	dogClient *animals.Client,
) error {
	dog, err := dogClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Despite how it looks it is well trained\n\n ![loyal soldier](%v)", dog.Url)
	err = createIssueComment(is, client, message)
	return err
}

//labelHandler handles the /label command
func labelHandler(
	is *github.IssueCommentEvent,
	client *paulclient.GithubClient,
	labels []string,
) error {
	_, _, err := client.IssueService.AddLabelsToIssue(
		client.Ctx,
		*is.Repo.Owner.Login,
		is.Repo.GetName(),
		is.Issue.GetNumber(),
		labels,
	)
	return err
}

//removeLabelHandler handles the /removelabel command
func removeLabelHandler(
	is *github.IssueCommentEvent,
	client *paulclient.GithubClient,
	label string,
) error {
	_, err := client.IssueService.RemoveLabelForIssue(
		client.Ctx,
		*is.Repo.Owner.Login,
		is.Repo.GetName(),
		is.Issue.GetNumber(),
		label,
	)
	return err
}

// createIssueComment sends a comment to an issue/pull request
func createIssueComment(
	is *github.IssueCommentEvent,
	client *paulclient.GithubClient,
	message string,
) error {
	comment := &github.IssueComment{Body: &message}
	_, _, err := client.IssueService.CreateComment(
		client.Ctx,
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
