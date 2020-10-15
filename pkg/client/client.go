package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// JWTAuth token issued by Github in response to signed JWT Token
type JWTAuth struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

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

type GithubClient struct {
	Ctx                context.Context
	GitService         gitService
	RepoService        repoService
	PullRequestService pullRequestService
	IssueService       issueService
}

//GetClient returns an authorized Github Client
func GetClient(installationIDString string) (*GithubClient, error) {
	installationID, err := strconv.ParseInt(installationIDString, 10, 64)
	if err != nil {
		return &GithubClient{}, err
	}
	ctx := context.Background()
	cfg, err := newConfig()
	if err != nil {
		return &GithubClient{}, err
	}
	token, err := getAccessToken(cfg, installationID)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	githubClient := &GithubClient{
		Ctx:                ctx,
		GitService:         client.Git,
		RepoService:        client.Repositories,
		PullRequestService: client.PullRequests,
		IssueService:       client.Issues,
	}
	return githubClient, err
}

// Config to run Paul
type config struct {
	SecretKey     string
	PrivateKey    string
	ApplicationID string
}

const (
	secretKeyFile  = "paul-secret-key"
	privateKeyFile = "paul-private-key"
)

// NewConfig populates configuration from known-locations and gives
// an error if configuration is missing from disk or environmental variables
func newConfig() (config, error) {
	config := config{}

	keyPath, pathErr := getSecretPath()
	if pathErr != nil {
		return config, pathErr
	}

	secretKeyBytes, readErr := ioutil.ReadFile(path.Join(keyPath, secretKeyFile))

	if readErr != nil {
		msg := fmt.Errorf("unable to read GitHub symmetrical secret: %s, error: %s",
			keyPath+secretKeyFile, readErr)
		return config, msg
	}

	secretKeyBytes = getFirstLine(secretKeyBytes)
	config.SecretKey = string(secretKeyBytes)

	privateKeyPath := path.Join(keyPath, privateKeyFile)

	keyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return config, fmt.Errorf("unable to read private key path: %s, error: %s", privateKeyPath, err)
	}

	config.PrivateKey = string(keyBytes)

	if val, ok := os.LookupEnv("APPLICATION_ID"); ok && len(val) > 0 {
		config.ApplicationID = val
	} else {
		return config, fmt.Errorf("APPLICATION_ID must be given")
	}

	return config, nil
}

func getSecretPath() (string, error) {
	secretPath := os.Getenv("SECRET_PATH")

	if len(secretPath) == 0 {
		return "", fmt.Errorf("SECRET_PATH env-var not set")
	}

	return secretPath, nil
}

func getFirstLine(secret []byte) []byte {
	stringSecret := string(secret)
	if newLine := strings.Index(stringSecret, "\n"); newLine != -1 {
		secret = secret[:newLine]
	}
	return secret
}

//GetAccessToken returns a Github OAuth Token
func getAccessToken(config config, installationID int64) (string, error) {
	token := os.Getenv("PERSONAL_ACCESS_TOKEN")
	if len(token) == 0 {
		installationToken, tokenErr := makeAccessTokenForInstallation(
			config.ApplicationID,
			installationID,
			config.PrivateKey,
		)
		if tokenErr != nil {
			return "", tokenErr
		}
		token = installationToken
	}
	return token, nil
}

// MakeAccessTokenForInstallation makes an access token for an installation / private key
func makeAccessTokenForInstallation(appID string, installation int64, privateKey string) (string, error) {
	signed, err := getSignedJwtToken(appID, privateKey)

	if err != nil {
		msg := fmt.Sprintf("can't run GetSignedJwtToken for app_id: %s and installation_id: %d, error: %v", appID, installation, err)

		fmt.Printf("Error %s\n", msg)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installation), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signed))
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		msg := fmt.Sprintf("can't get access_token for app_id: %s and installation_id: %d error: %v", appID, installation, err)
		fmt.Printf("Error: %s\n", msg)
		return "", fmt.Errorf("%s", msg)
	}

	defer res.Body.Close()

	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	jwtAuth := JWTAuth{}
	jsonErr := json.Unmarshal(bytesOut, &jwtAuth)
	if jsonErr != nil {
		return "", jsonErr
	}
	return jwtAuth.Token, nil
}

// GetSignedJwtToken get a tokens signed with private key
func getSignedJwtToken(appID string, privateKey string) (string, error) {

	keyBytes := []byte(privateKey)
	key, keyErr := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if keyErr != nil {
		return "", keyErr
	}

	now := time.Now()
	claims := jwt.StandardClaims{
		Issuer:    appID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 9).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedVal, signErr := token.SignedString(key)
	if signErr != nil {
		return "", signErr
	}
	return string(signedVal), nil
}
