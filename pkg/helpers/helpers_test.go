package helpers

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMockHTTPClient(t *testing.T) {
	t.Run("Test returns mock client", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`test`))
		})
		expected := &http.Client{}
		mockClient, close := MockHTTPClient(h)
		defer close()
		assert.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(mockClient))
	})
}
