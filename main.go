package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"

	"kuda/cmd/middleware"
)

type (
	Response struct {
		Message string `json:"message"`
		Date    string `json:"date"`
	}
)

var Logger *log.Logger

var (
	optPort = flag.Int("p", 8080, "port number")
)

func main() {
	Logger = log.New(os.Stdout, "kuda:", log.LstdFlags)

	ParseArgs()
	e := echo.New()
	Route(e)
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(*optPort)))
}

func ParseArgs() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: hwrap [flags] command\n")
		flag.PrintDefaults()
	}
	flag.Parse()
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

func genGCPIdToken(targetAudience string) string {
	authUrl := "https://www.googleapis.com/oauth2/v4/token"

	// ADC
	credentials, err := google.FindDefaultCredentials(oauth2.NoContext, authUrl)
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
		Aud: authUrl,
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
	res, err := http.PostForm(authUrl, f)
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

func Route(e *echo.Echo) {
	e.Use(middleware.Logger)

	e.GET("*", func(c echo.Context) (err error) {

		var hparams []string
		for k, v := range c.QueryParams() {
			hparams = append(hparams, k+"="+v[0])
		}

		client := &http.Client{}
		if err != nil {
			log.Fatal(err)
		}

		url := "https://kuda-target-dnb6froqha-uc.a.run.app/healthcheck"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header = c.Request().Header.Clone()
		req.Header.Add("X-Forwarded-For", c.Request().RemoteAddr)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", genGCPIdToken(url)))

		res, _ := client.Do(req)

		return c.Stream(res.StatusCode, res.Header.Get("Content-Type"), res.Body)
	})
}
