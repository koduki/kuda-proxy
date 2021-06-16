package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

const AuthUrl = "https://www.googleapis.com/oauth2/v4/token"

func IdToken(targetAudience string) string {

	// ADC
	credentials, err := google.FindDefaultCredentials(oauth2.NoContext, AuthUrl)
	config, err := google.JWTConfigFromJSON(credentials.JSON)
	if err != nil {
		fmt.Println(err)
	}

	// JWS
	iat := time.Now()
	exp := iat.Add(time.Hour)
	cs := &jws.ClaimSet{
		Iss: config.Email,
		Sub: config.Email,
		Aud: AuthUrl,
		Iat: iat.Unix(),
		Exp: exp.Unix(),
	}
	hdr := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		KeyID:     config.PrivateKeyID,
	}
	privateKey := ParseKey(config.PrivateKey)

	// Request Google OAuth server to get Token ID
	cs.PrivateClaims = map[string]interface{}{"target_audience": targetAudience}
	msg, err := jws.Encode(hdr, cs, privateKey)
	if err != nil {
		fmt.Println(fmt.Errorf("google: could not encode JWT: %v", err))
	}

	f := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {msg},
	}
	res, err := http.PostForm(AuthUrl, f)
	if err != nil {
		fmt.Println(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
	}

	type resIdToken struct {
		IdToken string `json:"id_token"`
	}
	id := &resIdToken{}
	json.Unmarshal(body, id)

	return id.IdToken
}

func ParseKey(key []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(key)
	if block != nil {
		key = block.Bytes
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		parsedKey, err = x509.ParsePKCS1PrivateKey(key)
		if err != nil {
			return nil
		}
	}
	parsed, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil
	}
	return parsed
}
