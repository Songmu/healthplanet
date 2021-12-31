package healthplanet

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const cmdName = "healthplanet"

// Run the healthplanet
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)

	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}
	app, err := newApp(ctx)
	if err != nil {
		return err
	}

	body := url.Values{}
	body.Set("from", "20211220000000")
	body.Set("to", "20220129000000")
	body.Set("tag", "6021,6022")
	resp, err := app.doAPI(ctx, "/status/innerscan.json", body)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println(string(b))

	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}

var dispatch = map[string]runner{
	"metrics": runnerFunc(metricsCmd),
	"request": runnerFunc(requestCmd),
}

type runner interface {
	run(context.Context, []string, io.Writer, io.Writer) error
}

type healthplanet struct {
	uri          *url.URL
	token        *oauth2.Token
	config       *oauth2.Config
	settingsFile string
}

const baseURL = "https://www.healthplanet.jp"

func newApp(ctx context.Context) (*healthplanet, error) {
	hp := &healthplanet{
		config: newOauth2Config(),
	}
	hp.uri, _ = url.Parse(baseURL)

	if err := hp.setup(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	if hp.token == nil || hp.token.AccessToken == "" {
		if err := hp.accessToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}
	if err := hp.refreshTokenIfInvalid(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return hp, nil
}

func (hp *healthplanet) setup() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir = filepath.Join(dir, "go-healthplanet")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	hp.settingsFile = filepath.Join(dir, "settings.json")
	f, err := os.Open(hp.settingsFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&hp.token); err != nil {
		return fmt.Errorf("could not unmarshal %s: %w", hp.settingsFile, err)
	}
	return nil
}

var defaultUserAgent string

func init() {
	defaultUserAgent = "Songmu/" + version + " (+https://github.com/Songmu/healthplanet)"
}

// Implement the token refresh logic on our own. This is because Healthplanet requires redirect_uri
// as a required parameter even for refresh requests, and `hp.config.TokenSource(ctx, hp.token).Token()`
// doesn't do it. (This is rather a strange behavior on Healthplanet's side.
func (hp *healthplanet) refreshTokenIfInvalid(ctx context.Context) error {
	if hp.isTokenValid() {
		return nil
	}
	log.Println("your token is expired, so refreshing")

	req, err := hp.refreshRequest(ctx)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if code := resp.StatusCode; code < 200 || code > 299 {
		return fmt.Errorf("oauth2: can not fetch token: %d\nResponse: %s", code, string(body))
	}
	var tj tokenJSON
	if err = json.Unmarshal(body, &tj); err != nil {
		return err
	}
	t := &oauth2.Token{
		AccessToken:  tj.AccessToken,
		TokenType:    tj.TokenType,
		RefreshToken: tj.RefreshToken,
		Expiry:       tj.expiry(),
	}
	hp.token = t
	if err := hp.saveToken(); err != nil {
		return fmt.Errorf("failed to saveToken: %w", err)
	}
	return nil
}

func expired(t *oauth2.Token) bool {
	if t.Expiry.IsZero() {
		return false
	}
	return t.Expiry.Round(0).Add(-10 * time.Second).Before(time.Now())
}

func (hp *healthplanet) isTokenValid() bool {
	t := hp.token
	return t != nil && t.AccessToken != "" && !expired(t)
}

func (hp *healthplanet) refreshRequest(ctx context.Context) (*http.Request, error) {
	v := url.Values{}
	v.Set("client_id", hp.config.ClientID)
	v.Set("client_secret", hp.config.ClientSecret)
	v.Set("redirect_uri", hp.config.RedirectURL)
	v.Set("grant_type", "refresh_token")
	v.Set("refresh_token", hp.token.RefreshToken)
	req, err := http.NewRequestWithContext(
		ctx, "POST", hp.config.Endpoint.TokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	return hp.setDefaultHeaders(req), nil
}

func (hp *healthplanet) setDefaultHeaders(req *http.Request) *http.Request {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", defaultUserAgent)
	return req
}

// almost copied from oauth2/token.go
// tokenJSON is the struct representing the HTTP response from OAuth2
// providers returning a token in JSON form.
type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (e *tokenJSON) expiry() (t time.Time) {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return
}

func (hp *healthplanet) doAPI(ctx context.Context, path string, body url.Values) (
	*http.Response, error) {

	if body == nil {
		body = url.Values{}
	}
	body.Set("access_token", hp.token.AccessToken)
	hp.uri.Path = path
	hp.uri.RawQuery = body.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", hp.uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req = hp.setDefaultHeaders(req)
	return http.DefaultClient.Do(req)
}
