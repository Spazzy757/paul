package client

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v49/github"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/oauth2"
)

const (
	secretKeyFile  = "paul-secret-key"
	privateKeyFile = "paul-private-key"
	githubBaseUrl  = "https://api.github.com"
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

// GetInstallationClient returns an authorized Github Client for an installation
func GetInstallationClient(installationID int64) (*github.Client, error) {
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

// GetClient returns an authorized Github Client
func GetClient() (*github.Client, error) {
	ctx := context.Background()
	cfg, err := newConfig()
	if err != nil {
		return &github.Client{}, err
	}
	signed, err := getSignedToken(cfg.ApplicationID, cfg.PrivateKey)
	if err != nil {
		return &github.Client{}, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: signed,
			TokenType:   "Bearer",
		},
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

	secretKeyBytes, readErr := os.ReadFile(path.Join(keyPath, secretKeyFile))

	if readErr != nil {
		msg := fmt.Errorf("unable to read GitHub symmetrical secret: %s, error: %s",
			keyPath+secretKeyFile, readErr)
		return config, msg
	}

	secretKeyBytes = getFirstLine(secretKeyBytes)
	config.SecretKey = string(secretKeyBytes)

	privateKeyPath := path.Join(keyPath, privateKeyFile)

	keyBytes, err := os.ReadFile(privateKeyPath)
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

// GetAccessToken returns a Github OAuth Token
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
	signed, err := getSignedToken(appID, privateKey)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(
		"%v/app/installations/%d/access_tokens",
		c.BaseUrl,
		installation,
	)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
	bytesOut, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	jwtAuth := JWTAuth{}
	jsonErr := json.Unmarshal(bytesOut, &jwtAuth)
	return jwtAuth.Token, jsonErr
}

// getSignedToken Returns a signed JWT Token
func getSignedToken(appID string, privateKey string) (string, error) {
	keyBytes := []byte(privateKey)
	pKey, err := bytesToPrivateKey(keyBytes)

	if err != nil {
		return "", err
	}
	realKey, err := jwk.New(pKey)
	if err != nil {
		return "", err
	}
	// Create the token
	token := jwt.New()

	now := time.Now()
	// Ignore errors of setting claims
	_ = token.Set(jwt.IssuerKey, appID)
	_ = token.Set(jwt.IssuedAtKey, now.Unix())
	_ = token.Set(jwt.ExpirationKey, now.Add(time.Minute*9).Unix())

	// Sign the token and generate a payload
	signed, err := jwt.Sign(token, jwa.RS256, realKey)
	if err != nil {
		log.Printf("HERE")
		return "", err
	}

	return string(signed), nil
}

// BytesToPrivateKey bytes to private key
func bytesToPrivateKey(priv []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(priv)
	if block == nil {
		return &rsa.PrivateKey{}, errors.New("Unable to Decode Private Key")
	}
	b := block.Bytes
	key, err := x509.ParsePKCS1PrivateKey(b)
	return key, err
}
