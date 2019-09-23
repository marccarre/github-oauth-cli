package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"os/exec"
	"strings"

	"github.com/marccarre/github-oauth-cli/pkg/random"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// clientID is eksctl's GitHub application client ID.
var clientID = os.Getenv("CLIENT_ID")

// clientSecret is eksctl's GitHub application client secret.
var clientSecret = os.Getenv("CLIENT_SECRET")

type authorizeReq struct {
	clientID    string // client_id: Required. The client ID you received from GitHub when you registered.
	redirectURI string // redirect_uri: The URL in your application where users will be sent after authorization.
	login       string // login: Suggests a specific account to use for signing in and authorizing the app.
	scope       string // scope: A space-delimited list of scopes. If not provided, scope defaults to an empty list for users that have not authorized any scopes for the application. For users who have authorized scopes for the application, the user won't be shown the OAuth authorization page with the list of scopes. Instead, this step of the flow will automatically complete with the set of scopes the user has authorized for the application.
	state       string // state: An unguessable random string. It is used to protect against cross-site request forgery attacks.
}

func (r authorizeReq) rawRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/login/oauth/authorize", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise authorize request")
	}
	q := req.URL.Query()
	q.Add("client_id", r.clientID)
	q.Add("redirect_uri", r.redirectURI)
	q.Add("login", r.login)
	q.Add("scope", r.scope)
	q.Add("state", r.state)
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (r authorizeReq) Send(ctx context.Context) (string, error) {
	req, err := r.rawRequest(ctx)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{"url": req.URL.String()}).Info("URL-encoded authorize request")

	client := http.Client{}
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "authorize request failed")
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", errors.Wrap(err, "failed to read authorize response")
	}
	return string(body), nil
}

type accessTokenReq struct {
	clientID     string // client_id: Required. The client ID you received from GitHub for your GitHub App.
	clientSecret string // client_secret: Required. The client secret you received from GitHub for your GitHub App.
	code         string // code: Required. The code you received as a response to Step 1. (authorizeReq)
	redirectURI  string // redirect_uri: The URL in your application where users are sent after authorization.
	state        string // state: The unguessable random string you provided in Step 1. (authorizeReq)
}

func (r accessTokenReq) Send(ctx context.Context) (*accessTokenRes, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", bytes.NewBufferString(url.Values{
		"client_id":     {r.clientID},
		"client_secret": {r.clientSecret},
		"code":          {r.code},
		"redirect_uri":  {r.redirectURI},
		"state":         {r.state},
	}.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise access token request")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := http.Client{}
	defer client.CloseIdleConnections()
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "access token request failed")
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read access token response")
	}
	var data accessTokenRes
	if err := data.Unmarshal(body); err != nil {
		return nil, errors.Wrap(err, "failed to deserialise access token response")
	}
	log.WithFields(log.Fields{"data": data}).Info("access token succeeded")
	return &data, nil
}

type accessTokenRes struct {
	accessToken string // `json:"access_token"`
	rawScope    string // `json:"scope"`
	scopes      []string
	tokenType   string // `json:"token_type"`
}

func (r *accessTokenRes) Unmarshal(bytes []byte) error {
	if err := json.Unmarshal(bytes, r); err != nil {
		return err
	}
	r.scopes = strings.Split(r.rawScope, ",")
	return nil
}

// GetToken gets a GitHub OAuth token via a web flow.
func GetToken(ctx context.Context, username string) (*oauth2.Token, error) {
	channel := make(chan string)
	state := random.String(20)
	authServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != state {
			log.Errorf("state doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			channel <- code
			return
		}
		log.Errorf("No code")
		http.Error(rw, "", 500)
	}))
	defer authServer.Close()

	authReq := authorizeReq{
		clientID:    clientID,
		redirectURI: authServer.URL,
		login:       username,
		scope:       "repo",
		state:       state,
	}
	rawReq, err := authReq.rawRequest(ctx)
	if err != nil {
		return nil, err
	}
	authURL := rawReq.URL.String()
	go openURL(authURL)
	log.Infof("Please authorise this app at: %s", authURL)
	code := <-channel
	log.Infof("Got code: %s", code)

	tokenReq := accessTokenReq{
		clientID:     clientID,
		clientSecret: clientSecret,
		code:         code,
		state:        state,
	}
	tokenRep, err := tokenReq.Send(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get OAuth token")
	}
	return &oauth2.Token{AccessToken: tokenRep.accessToken}, nil
}

func openURL(url string) {
	bins := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range bins {
		err := exec.Command(bin, url).Run() // #nosec: URL is based on constant google.Endpoint
		if err == nil {
			return
		}
	}
	log.WithField("url", url).Errorf("Error opening URL in browser")
}
