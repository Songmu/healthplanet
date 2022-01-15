package healthplanet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL *url.URL
	token   string

	client *http.Client
}

func newClient(u *url.URL, token string) *Client {
	return &Client{
		baseURL: u,
		token:   token,
		client:  http.DefaultClient,
	}
}

var st2tags = map[string]string{
	"innerscan":        "6021,6022",      // kg, %
	"sphygmomanometer": "622E,622F,6230", // mmHg, mmHg, bpm
	"pedometer":        "6331",           // walk
}

func (cl *Client) getStatus(ctx context.Context, status string, from, to time.Time) (*response, error) {
	const layout = "20060102150405"

	endpoint := fmt.Sprintf("/status/%s.json", status)
	tags, ok := st2tags[status]
	if !ok {
		return nil, fmt.Errorf("no tag found for status: %s", status)
	}
	body := url.Values{}
	body.Set("from", from.Format(layout))
	body.Set("to", to.Format(layout))
	body.Set("date", "1")
	body.Set("tag", tags)
	resp, err := cl.doAPI(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d := &response{}
	if err := json.NewDecoder(resp.Body).Decode(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (cl *Client) doAPI(ctx context.Context, path string, body url.Values) (*http.Response, error) {
	u := *cl.baseURL
	if body == nil {
		body = url.Values{}
	}
	body.Set("access_token", cl.token)
	u.Path = path
	u.RawQuery = body.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return cl.do(req)
}

var defaultUserAgent string

func init() {
	defaultUserAgent = "Songmu/" + version + " (+https://github.com/Songmu/healthplanet)"
}

func (cl *Client) setDefaultHeaders(req *http.Request) *http.Request {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", defaultUserAgent)
	return req
}

func (cl *Client) do(req *http.Request) (*http.Response, error) {
	req = cl.setDefaultHeaders(req)
	return cl.client.Do(req)
}
