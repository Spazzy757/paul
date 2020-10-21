package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

type mockRepoClient struct {
	resp io.ReadCloser
}

func (m *mockRepoClient) DownloadContents(
	ctx context.Context,
	owner, repo, filepath string,
	opt *github.RepositoryContentGetOptions,
) (io.ReadCloser, error) {
	return m.resp, nil
}

func TestNewConfigNoSecretPath(t *testing.T) {

	os.Setenv("SECRET_PATH", "")

	_, err := newConfig()

	if err == nil {
		t.Fail()
	}

	want := "SECRET_PATH env-var not set"
	if err.Error() != want {
		t.Errorf("want %q, got %q", want, err.Error())
		t.Fail()
	}
}

func TestNewConfigValidSecretPathWithApplicationID(t *testing.T) {
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

	cfg, err := newConfig()

	if err != nil {
		t.Errorf("%s", err.Error())
		t.Fail()
		return
	}

	if cfg.SecretKey != secretWant {
		t.Errorf("want %q, got %q", secretWant, cfg.SecretKey)
		t.Fail()
	}

	if cfg.PrivateKey != privateWant {
		t.Errorf("want %q, got %q", privateWant, cfg.PrivateKey)
		t.Fail()
	}

	if cfg.ApplicationID != appIDWant {
		t.Errorf("want %q, got %q", appIDWant, cfg.ApplicationID)
		t.Fail()
	}
}

func TestGetFirstLine(t *testing.T) {
	var exampleSecrets = []struct {
		secret       string
		expectedByte string
	}{
		{
			secret:       "New-line \n",
			expectedByte: "New-line ",
		},
		{
			secret: `Newline and text 
			`,
			expectedByte: "Newline and text ",
		},
		{
			secret:       `Example secret2 `,
			expectedByte: `Example secret2 `,
		},
		{
			secret:       "\n",
			expectedByte: "",
		},
		{
			secret:       "",
			expectedByte: "",
		},
	}
	for _, test := range exampleSecrets {

		t.Run(string(test.secret), func(t *testing.T) {
			stringNoLines := getFirstLine([]byte(test.secret))
			if test.expectedByte != string(stringNoLines) {
				t.Errorf("String after removal - wanted: \"%s\", got \"%s\"", test.expectedByte, test.secret)
			}
		})
	}
}

func TestAccessToken(t *testing.T) {
	serverUrl, _, teardown := ServerMock()
	defer teardown()
	aClient := &authClient{
		BaseUrl: serverUrl,
		Client:  http.DefaultClient,
	}

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

		cfg, _ := newConfig()

		token, _ := getAccessToken(aClient, cfg, 1)
		assert.Equal(t, token, "123456789")
	})
	t.Run("Test Unset Environment Err", func(t *testing.T) {
		os.Unsetenv("PERSONAL_ACCESS_TOKEN")

		cfg, _ := newConfig()

		token, _ := getAccessToken(aClient, cfg, 1)
		assert.Equal(t, token, "")
	})
}

func TestSignedJWTToken(t *testing.T) {
	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)

	publicKey := key.PublicKey
	pemSecretfile, err := os.Create(path.Join(tmpDir, "privatekey.pem"))
	assert.Equal(t, err, nil)
	defer pemSecretfile.Close()

	err = pem.Encode(
		pemSecretfile,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	assert.Equal(t, err, nil)
	pemPublicfile, err := os.Create(path.Join(tmpDir, "pubKey.pem"))
	assert.Equal(t, err, nil)
	defer pemPublicfile.Close()

	asn1Bytes, _ := x509.MarshalPKIXPublicKey(&publicKey)
	assert.Equal(t, err, nil)
	_ = pem.Encode(
		pemPublicfile,
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: asn1Bytes,
		},
	)

	defer os.RemoveAll(path.Join(tmpDir, "privatekey.pem"))
	defer os.RemoveAll(path.Join(tmpDir, "pubkey.pem"))

	pubKeyReader, _ := ioutil.ReadFile(path.Join(tmpDir, "pubKey.pem"))
	signingKey, _ := ioutil.ReadFile(path.Join(tmpDir, "privatekey.pem"))
	t.Run("Test Getting a Singed Token", func(t *testing.T) {
		tokenString, err := getSignedJwtToken("123456789", string(signingKey))
		assert.Equal(t, err, nil)
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyReader)
		assert.Equal(t, err, nil)
		claims := &jwt.StandardClaims{}
		keyFn := func(t *jwt.Token) (interface{}, error) { return &pubKey, nil }
		token, _ := jwt.ParseWithClaims(tokenString, claims, keyFn)
		claims, _ = token.Claims.(*jwt.StandardClaims)
		assert.Equal(t, "123456789", claims.Issuer)
	})
}

func ServerMock() (baseURL string, mux *http.ServeMux, teardownFn func()) {
	mux = http.NewServeMux()
	srv := httptest.NewServer(mux)
	return srv.URL, mux, srv.Close
}

func TestMakeAccessTokenForInstallation(t *testing.T) {
	serverUrl, mux, teardown := ServerMock()
	defer teardown()
	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)

	pemSecretfile, _ := os.Create(path.Join(tmpDir, "privatekey.pem"))
	defer pemSecretfile.Close()

	_ = pem.Encode(
		pemSecretfile,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	signingKey, _ := ioutil.ReadFile(path.Join(tmpDir, "privatekey.pem"))

	defer os.RemoveAll(path.Join(tmpDir, "privatekey.pem"))
	t.Run("Test Getting Token For Installation", func(t *testing.T) {
		aClient := &authClient{
			BaseUrl: serverUrl,
			Client:  http.DefaultClient,
		}
		mux.HandleFunc(
			"/app/installations/645/access_tokens",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "POST")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"token":"123456"}`)
			},
		)
		token, err := makeAccessTokenForInstallation(aClient, "123", 645, string(signingKey))
		assert.Equal(t, nil, err)
		assert.Equal(t, "123456", token)
	})
	t.Run("Test Getting Token For Installation", func(t *testing.T) {
		aClient := &authClient{
			BaseUrl: serverUrl,
			Client:  http.DefaultClient,
		}
		mux.HandleFunc(
			"/app/installations/644/access_tokens",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "POST")
				w.WriteHeader(http.StatusInternalServerError)
			},
		)
		_, err := makeAccessTokenForInstallation(aClient, "123", 644, string(signingKey))
		assert.NotEqual(t, nil, err)
	})

}
