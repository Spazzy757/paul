package animals

import (
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClientGetCat(t *testing.T) {
	catAPIResponse := `[
        {
            "breeds":[],
            "id":"40g",
            "url":"https://cdn2.thecatapi.com/images/40g.jpg",
            "width":640,
            "height":426
         }
    ]`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(catAPIResponse))
	})
	httpClient, teardown := helpers.MockHTTPClient(h)
	defer teardown()

	client := NewCatClient()
	client.HttpClient = httpClient
	assert.Equal(t, client.Url, "https://api.thecatapi.com/v1/images/search")
	client.Url = "https://example.com"
	t.Run("Test Get Cat Returns Cat Type", func(t *testing.T) {
		cat, err := client.GetLink()
		assert.Nil(t, err)
		assert.Equal(t, cat.Url, "https://cdn2.thecatapi.com/images/40g.jpg")
	})
}

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
	assert.Equal(t, client.Url, "https://api.thedogapi.com/v1/images/search")
	client.Url = "https://example.com"
	t.Run("Test Get Dog Returns Dog Type", func(t *testing.T) {
		dog, err := client.GetLink()
		assert.Nil(t, err)
		assert.Equal(t, dog.Url, "https://cdn2.thedogapi.com/images/ryHJZlcNX_1280.jpg")
	})
}

func TestAnimalClient(t *testing.T) {

	client := NewDogClient()
	t.Run("Test Error on Upstream Server", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMovedPermanently)
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		client.HttpClient = httpClient
		client.Url = "https://example.com"
		_, err := client.GetLink()
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test Error UnMarshaling Json", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`test`))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		client.HttpClient = httpClient
		client.Url = "https://example.com"
		_, err := client.GetLink()
		assert.NotEqual(t, nil, err)
	})
}
