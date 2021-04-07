package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/google/go-github/v33/github"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
)

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
func TestGetClient(t *testing.T) {
	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)
	privateWant := "private"
	_ = ioutil.WriteFile(path.Join(tmpDir, "paul-secret-key"), []byte(privateWant), 0600)
	pemSecretfile, err := os.Create(path.Join(tmpDir, "paul-private-key"))
	assert.Equal(t, err, nil)
	defer pemSecretfile.Close()

	err = pem.Encode(
		pemSecretfile,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	assert.Equal(t, nil, err)

	defer os.RemoveAll(path.Join(tmpDir, "paul-private-key"))
	defer os.RemoveAll(path.Join(tmpDir, "paul-secret-key"))
	os.Setenv("SECRET_PATH", tmpDir)
	os.Setenv("APPLICATION_ID", "1234")
	t.Run("Test returns client with no error", func(t *testing.T) {
		os.Setenv("SECRET_PATH", tmpDir)
		os.Setenv("APPLICATION_ID", "1234")
		client, err := GetClient()
		expected := &github.Client{}
		assert.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(client))
		assert.Equal(t, nil, err)
	})
	t.Run("Test returns client and error", func(t *testing.T) {
		os.Setenv("SECRET_PATH", "/doesnt/exists")
		os.Setenv("APPLICATION_ID", "1234")
		client, err := GetClient()
		expected := &github.Client{}
		assert.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(client))
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test returns client and error when key is not valid", func(t *testing.T) {
		privateWant := "private"
		secretWant := "secret"
		appIDWant := "321"
		tmpDir := os.TempDir()

		_ = ioutil.WriteFile(path.Join(tmpDir, "paul-private-key"), []byte(privateWant), 0600)
		_ = ioutil.WriteFile(path.Join(tmpDir, "paul-secret-key"), []byte(secretWant), 0600)

		defer os.RemoveAll(path.Join(tmpDir, "paul-private-key"))
		defer os.RemoveAll(path.Join(tmpDir, "paul-secret-key"))

		os.Setenv("SECRET_PATH", tmpDir)
		os.Setenv("APPLICATION_ID", appIDWant)
		client, err := GetClient()
		expected := &github.Client{}
		assert.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(client))
		assert.NotEqual(t, nil, err)
	})

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
			expectedByte: "Newline and text",
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
			assert.Equal(t, test.expectedByte, string(stringNoLines))
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

	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)

	pemSecretfile, err := os.Create(path.Join(tmpDir, "paul-private-key"))
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

	defer os.RemoveAll(path.Join(tmpDir, "paul-private-key"))

	secretWant := "secret"
	appIDWant := "321"

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

func TestGetSignedToken(t *testing.T) {
	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)

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

	defer os.RemoveAll(path.Join(tmpDir, "privatekey.pem"))

	signingKey, _ := ioutil.ReadFile(path.Join(tmpDir, "privatekey.pem"))
	t.Run("Test Getting a Singed Token", func(t *testing.T) {
		tokenString, err := getSignedToken("123456789", string(signingKey))
		assert.Equal(t, err, nil)
		token, err := jwt.Parse(
			[]byte(tokenString),
			jwt.WithValidate(true),
			jwt.WithVerify(jwa.RS256, &key.PublicKey),
		)
		assert.Equal(t, err, nil)
		assert.Equal(t, "123456789", token.Issuer())
	})
	t.Run("Test Getting a Singed Token", func(t *testing.T) {
		_, err := getSignedToken("123456789", "invalidKey")
		assert.NotEqual(t, err, nil)
	})
	t.Run("Test Bad RSA key", func(t *testing.T) {
		tmpDir := os.TempDir()
		reader := rand.Reader
		bitSize := 32
		key, err := rsa.GenerateKey(reader, bitSize)
		assert.Equal(t, err, nil)

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

		defer os.RemoveAll(path.Join(tmpDir, "privatekey.pem"))

		signingKey, _ := ioutil.ReadFile(path.Join(tmpDir, "privatekey.pem"))
		_, err = getSignedToken("123456789", string(signingKey))
		assert.NotEqual(t, err, nil)
	})
}
