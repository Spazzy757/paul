package github

import (
	"fmt"
	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/google/go-github/v32/github"
	"strings"
)

/*
IssueCommentHandler takes an incoming event of type IssueCommentEvent and
runs logic against it
*/
func IssueCommentHandler(event *github.IssueCommentEvent) error {
	// load github client
	client, ctx, clientErr := getClient(*event.Installation.ID)
	if clientErr != nil {
		return clientErr
	}
	// load Paul Config from repo
	rc := &repoClient{ctx: ctx, repoService: client.Repositories}
	cfg, configErr := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		event.Repo.GetDefaultBranch(),
		rc,
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
		// Create Client To pass through to handlers
		isClient := &issueClient{
			ctx:          ctx,
			issueService: client.Issues,
		}
		// Switch statement to handle different commands
		switch {
		// Case of /cat command
		case cmd == "cat" && cfg.PullRequests.CatsEnabled:
			// Get the Cat Client
			animalClient := animals.NewCatClient()
			err = catsHandler(event, isClient, animalClient)
		// Case of /dog command
		case cmd == "dog" && cfg.PullRequests.DogsEnabled:
			// Get the Dog Client
			animalClient := animals.NewDogClient()
			err = dogsHandler(event, isClient, animalClient)
		// Case /label command
		case cmd == "label" &&
			cfg.Labels &&
			checkStringInList(cfg.Maintainers, *event.Sender.Login):
			// handle the labels
			err = labelHandler(event, isClient, args)
		// Case /remove-label command
		case cmd == "remove-label" &&
			cfg.Labels &&
			checkStringInList(cfg.Maintainers, *event.Sender.Login):
			// handle the remove labels,
			// if more than one arg is passed through, don't do anything
			if len(args) == 1 {
				err = removeLabelHandler(event, isClient, args[0])
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
	isClient *issueClient,
	catClient *animals.Client,
) error {
	cat, err := catClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("My Most Trusted Minion\n\n ![my favorite minion](%v)", cat.Url)
	err = createIssueComment(is, isClient, message)
	return err
}

// handleDogs is the handler for the /dog command
func dogsHandler(
	is *github.IssueCommentEvent,
	isClient *issueClient,
	dogClient *animals.Client,
) error {
	dog, err := dogClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Despite how it looks it is well trained\n\n ![loyal soldier](%v)", dog.Url)
	err = createIssueComment(is, isClient, message)
	return err
}

//labelHandler handles the /label command
func labelHandler(
	is *github.IssueCommentEvent,
	isClient *issueClient,
	labels []string,
) error {
	_, _, err := isClient.issueService.AddLabelsToIssue(
		isClient.ctx,
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
	isClient *issueClient,
	label string,
) error {
	_, err := isClient.issueService.RemoveLabelForIssue(
		isClient.ctx,
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
	client *issueClient,
	message string,
) error {
	comment := &github.IssueComment{Body: &message}
	_, _, err := client.issueService.CreateComment(
		client.ctx,
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
