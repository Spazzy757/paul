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
	client.Url = "https://example.com"
	t.Run("Test Get Cat Returns Cat Type", func(t *testing.T) {
		cat, err := client.GetCat()
		assert.Nil(t, err)
		assert.Equal(t, cat.Url, "https://cdn2.thecatapi.com/images/40g.jpg")
	})
}
