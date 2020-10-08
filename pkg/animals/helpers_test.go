package animals

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("Test Get Cat Returns Cat Type", func(t *testing.T) {
		client := newClient("example.com")
		expected := &Client{}
		assert.Equal(t, reflect.TypeOf(client), reflect.TypeOf(expected))
	})
}
