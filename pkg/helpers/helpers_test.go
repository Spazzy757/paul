package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/stretchr/testify/assert"

	jwt "github.com/dgrijalva/jwt-go"
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
