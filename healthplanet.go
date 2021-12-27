package healthplanet

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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
	_, err := newApp(ctx, outStream, errStream)
	if err != nil {
		return err
	}
	return nil
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}

type healthplanet struct {
	outStream, errStream io.Writer

	token        *oauth2.Token
	config       *oauth2.Config
	settingsFile string
}

func newApp(ctx context.Context, outStream, errStream io.Writer) (*healthplanet, error) {
	hp := &healthplanet{
		config:    newOauth2Config(),
		outStream: outStream,
		errStream: errStream,
	}
	if err := hp.setup(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	if hp.token == nil || hp.token.AccessToken == "" {
		if err := hp.accessToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}
	t, err := hp.config.TokenSource(ctx, hp.token).Token()
	if err != nil {
		return nil, err
	}
	if t.AccessToken != hp.token.AccessToken {
		hp.token = t
		if err := hp.saveToken(); err != nil {
			return nil, err
		}
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
