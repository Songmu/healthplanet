package healthplanet

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Songmu/prompter"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

func newOauth2Config() *oauth2.Config {
	// Since this is just a client application, it doesn't matter if the secret is exposed.
	return &oauth2.Config{
		Scopes: []string{"innerscan,sphygmomanometer,pedometer"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.healthplanet.jp/oauth/auth",
			TokenURL: "https://www.healthplanet.jp/oauth/token",
		},
		ClientID:     "2524.ztEIB5uORk.apps.healthplanet.jp",
		ClientSecret: "1640611990960-13isYCizEIVHtxa9htPrYAe3cCy0FAaJtuu1UkbN",
		RedirectURL:  "https://www.healthplanet.jp/success.html",
	}
}

func (hp *healthplanet) accessToken(ctx context.Context) (err error) {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return err
	}
	state := fmt.Sprintf("%x", stateBytes)
	uri := hp.config.AuthCodeURL(state, oauth2.SetAuthURLParam("response_type", "code"))

	log.Printf("opening browser: %s\n", uri)
	log.Printf("The above URL will be opened to obtain an access token.")
	log.Printf("(If your browser does not open it automatically, please open it yourself.)\n")
	log.Printf("Go through the authorization flow in your browser, copy the code to get the token,\n")
	log.Printf("and then come back to this terminal and paste the code.")
	if !prompter.YN("Are you ready?", false) {
		return fmt.Errorf("token obtaining process has been canceled.")
	}
	if err := open.Start(uri); err != nil {
		return err
	}

	code := prompter.Prompt("enter the code", "")
	hp.token, err = hp.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange access token: %w", err)
	}
	return hp.saveToken()
}

func (hp *healthplanet) saveToken() (err error) {
	f, err := os.OpenFile(hp.settingsFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	jenc := json.NewEncoder(f)
	jenc.SetIndent("", "  ")
	if err := jenc.Encode(hp.token); err != nil {
		return fmt.Errorf("failed to store file: %v", err)
	}
	return nil
}
