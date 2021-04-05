package gif

import (
	"net/http"
	"os"
	"testing"

	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetGif(t *testing.T) {
	os.Setenv("GIPHY_API_KEY", "123456789")
	giphyResponse := `
    {
   	 	"data": [
			{
				"images":
				{
					"fixed_height":
					{
						"height": "200",
						"width": "356",
						"size": "319732",
						"url": "https://media1.giphy.com/media/iXQ8SgaMQAgtq/200.gif?cid=479f44c89j1oe6ka1wdran4m31ljfqx6scvrqbcj08ly81iq&rid=200.gif",
						"mp4_size": "55870",
						"mp4": "https://media1.giphy.com/media/iXQ8SgaMQAgtq/200.mp4?cid=479f44c89j1oe6ka1wdran4m31ljfqx6scvrqbcj08ly81iq&rid=200.mp4",
						"webp_size": "156666",
						"webp": "https://media1.giphy.com/media/iXQ8SgaMQAgtq/200.webp?cid=479f44c89j1oe6ka1wdran4m31ljfqx6scvrqbcj08ly81iq&rid=200.webp"
					}
				}
			}
		]
	}
    `
	client := NewGifClient()
	t.Run("Test Get Giphy", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(giphyResponse))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()
		client.HttpClient = httpClient
		client.Url = "https://example.com"

		gifUrl, err := client.GetLink("LGTM")
		assert.Equal(t, err, nil)
		assert.Equal(t, gifUrl, "https://media1.giphy.com/media/iXQ8SgaMQAgtq/200.gif?cid=479f44c89j1oe6ka1wdran4m31ljfqx6scvrqbcj08ly81iq&rid=200.gif")
	})
	t.Run("Test Error UnMarshaling Json", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`test`))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		client.HttpClient = httpClient
		client.Url = "https://example.com"
		_, err := client.GetLink("LGTM")
		assert.NotEqual(t, nil, err)
	})
}
