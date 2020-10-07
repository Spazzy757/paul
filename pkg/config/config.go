package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	secretKeyFile  = "paul-secret-key"
	privateKeyFile = "paul-private-key"
)

// Config to run Derek
type Config struct {
	SecretKey     string
	PrivateKey    string
	ApplicationID string
}

// NewConfig populates configuration from known-locations and gives
// an error if configuration is missing from disk or environmental variables
func NewConfig() (Config, error) {
	config := Config{}

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
