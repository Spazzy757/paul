package helpers

import (
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("SET_ENV", "1")
	t.Run("Test Unset Environment Returns Default", func(t *testing.T) {
		environment := GetEnv("UNSET_ENV", "default")
		assert.Equal(t, environment, "default")
	})
	t.Run("Test Set Environment Returns Default", func(t *testing.T) {
		environment := GetEnv("SET_ENV", "2")
		assert.Equal(t, environment, "1")
	})
}

func TestAccessToken(t *testing.T) {
	privateWant := "private"
	secretWant := "secret"
	appIDWant := "321"
	tmpDir := os.TempDir()

	ioutil.WriteFile(path.Join(tmpDir, "paul-private-key"), []byte(privateWant), 0600)
	ioutil.WriteFile(path.Join(tmpDir, "paul-secret-key"), []byte(secretWant), 0600)

	defer os.RemoveAll(path.Join(tmpDir, "paul-private-key"))
	defer os.RemoveAll(path.Join(tmpDir, "paul-secret-key"))
	os.Setenv("SECRET_PATH", tmpDir)
	os.Setenv("APPLICATION_ID", appIDWant)
	t.Run("Test set Environment Returns Personal Token token", func(t *testing.T) {
		os.Setenv("PERSONAL_ACCESS_TOKEN", "123456789")

		cfg, _ := config.NewConfig()

		token, _ := GetAccessToken(cfg, 1)
		assert.Equal(t, token, "123456789")
	})
	t.Run("Test Unset Environment Err", func(t *testing.T) {
		os.Unsetenv("PERSONAL_ACCESS_TOKEN")

		cfg, _ := config.NewConfig()

		token, _ := GetAccessToken(cfg, 1)
		assert.Equal(t, token, "")
	})
}
