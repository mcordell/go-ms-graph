package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

// AuthorizationCode is a value provided after initial successful
// authentication/authorization, it is used to get access/refresh tokens
type AuthorizationCode struct {
	Value string
}

// Tokens holds access and refresh tokens
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthorizationURL is the endpoint used for intial login/auth
const AuthorizationURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"

// TokenURL is the endpoint for getting access/refresh tokens
const TokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"

// GetTokens retrieves access and refresh tokens for a given scope
func GetTokens(c AuthorizationConfig, authCode AuthorizationCode, scope string) (t Tokens, err error) {
	formVals := url.Values{}
	formVals.Set("code", authCode.Value)
	formVals.Set("grant_type", "authorization_code")
	formVals.Set("redirect_uri", c.RedirectURL())
	formVals.Set("scope", scope)
	if c.ClientSecret != "" {
		formVals.Set("client_secret", c.ClientSecret)
	}
	formVals.Set("client_id", c.ClientID)
	response, err := http.PostForm(TokenURL, formVals)

	if err != nil {
		return t, errors.Wrap(err, "error while trying to get tokens")
	}
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return t, errors.Wrap(err, "error while trying to read token json body")
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to parse token json body")
	}

	return
}

// startLocalListener opens an http server to retrieve the redirect from initial
// authentication and set the authorization code's value
func startLocalListener(c AuthorizationConfig, token *AuthorizationCode) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%s", c.RedirectPort)}

	http.HandleFunc(c.RedirectPath, func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Fatalf("Error while parsing form from response %s", err)
			return
		}
		for k, v := range r.Form {
			if k == "code" {
				token.Value = strings.Join(v, "")
			}
		}

		fmt.Fprintf(w, "Auth done, you can close this window")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

// LoginRequest asks the os to open the login URL and starts a listening on the
// configured port for the authorizaton code. This is used on initial login to
// get the initial token pairs
func LoginRequest(c AuthorizationConfig) (token AuthorizationCode) {
	formVals := url.Values{}
	formVals.Set("grant_type", "authorization_code")
	formVals.Set("redirect_uri", c.RedirectURL())
	formVals.Set("scope", c.Scope)
	formVals.Set("response_type", "code")
	formVals.Set("response_mode", "query")
	formVals.Set("client_id", c.ClientID)
	uri, _ := url.Parse(AuthorizationURL)
	uri.RawQuery = formVals.Encode()

	cmd := exec.Command(c.OpenCMD, uri.String())
	err := cmd.Start()
	if err != nil {
		panic(errors.Wrap(err, "Error while opening login URL"))

	}
	running := true
	srv := startLocalListener(c, &token)
	for running {
		if token.Value != "" {
			if err := srv.Shutdown(context.TODO()); err != nil {
				panic(err) // failure/timeout shutting down the server gracefully
			}
			running = false
		}
	}
	return
}
