package healthplanet

import "golang.org/x/oauth2"

func newConfig() *oauth2.Config {
	return &oauth2.Config{
		Scopes: []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.healthplanet.jp/oauth/auth",
			TokenURL: "https://www.healthplanet.jp/oauth/token",
		},
		ClientID:     "2522.NBNCEddrU4.apps.healthplanet.jp",
		ClientSecret: "1640595736253-xQz56WhKspgfHDZZ5sdhPQgqiAeICgWohbuKTYKW",
		RedirectURL:  "http://localhost:9545",
	}
}
