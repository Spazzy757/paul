package animals

import (
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClientGetDog(t *testing.T) {
	dogAPIResponse := `[
        {
            "breeds":[],
            "id":"ryHJZlcNX",
            "url":"https://cdn2.thedogapi.com/images/ryHJZlcNX_1280.jpg",
            "width":577,
            "height":634
        }
    ]`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(dogAPIResponse))
	})
	httpClient, teardown := helpers.MockHTTPClient(h)
	defer teardown()

	client := NewDogClient()
	client.HttpClient = httpClient
	client.Url = "https://example.com"
	t.Run("Test Get Dog Returns Dog Type", func(t *testing.T) {
		dog, err := client.GetCat()
		assert.Nil(t, err)
		assert.Equal(t, dog.Url, "https://cdn2.thedogapi.com/images/ryHJZlcNX_1280.jpg")
	})
}
