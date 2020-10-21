package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

const (
	secretKeyFile  = "paul-secret-key"
	privateKeyFile = "paul-private-key"
	githubBaseUrl  = "https://api.github.com/"
)

// JWTAuth token issued by Github in response to signed JWT Token
type JWTAuth struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Config to run Paul
type config struct {
	SecretKey     string
	PrivateKey    string
	ApplicationID string
}

type authClient struct {
	BaseUrl string
	Client  *http.Client
	Ctx     context.Context
}

//GetClient returns an authorized Github Client
func GetClient(installationID int64) (*github.Client, error) {
	ctx := context.Background()
	cfg, err := newConfig()
	if err != nil {
		return &github.Client{}, err
	}
	aClient := &authClient{
		BaseUrl: githubBaseUrl,
		Client:  http.DefaultClient,
	}
	token, err := getAccessToken(aClient, cfg, installationID)
	if err != nil {
		return &github.Client{}, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client, err
}

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
func getAccessToken(client *authClient, config config, installationID int64) (string, error) {
	token := os.Getenv("PERSONAL_ACCESS_TOKEN")
	if len(token) == 0 {
		installationToken, tokenErr := makeAccessTokenForInstallation(
			client,
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
func makeAccessTokenForInstallation(
	c *authClient,
	appID string,
	installation int64,
	privateKey string,
) (string, error) {
	signed, err := getSignedJwtToken(appID, privateKey)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			"%v/app/installations/%d/access_tokens",
			c.BaseUrl,
			installation,
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signed))
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	res, err := c.Client.Do(req)

	if err != nil {
		return "", fmt.Errorf("error getting Access token %v", err)
	}

	defer res.Body.Close()

	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	jwtAuth := JWTAuth{}
	jsonErr := json.Unmarshal(bytesOut, &jwtAuth)
	return jwtAuth.Token, jsonErr
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
