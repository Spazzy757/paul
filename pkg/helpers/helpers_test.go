package helpers

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("SET_ENV", "1")
	t.Run("Test Unset Environment Returns Default", func(t *testing.T) {
		environment := GetEnv("UNSET_ENV", "default")
		if environment != "default" {
			t.Errorf("environment = %v; want default", environment)
		}
	})
	t.Run("Test Set Environment Returns Default", func(t *testing.T) {
		environment := GetEnv("SET_ENV", "2")
		if environment != "1" {
			t.Errorf("environment = %v; want 1", environment)
		}
	})
}
