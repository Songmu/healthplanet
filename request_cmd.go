package healthplanet

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"time"
)

func requestCmd(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	fs := flag.NewFlagSet("healthplanet request", flag.ContinueOnError)
	fs.SetOutput(errStream)
	var (
		status string
	)
	fs.StringVar(&status, "status", "innerscan", "kind of status")
	// TODO: from, to
	if err := fs.Parse(argv); err != nil {
		return err
	}

	app := getApp(ctx)
	now := time.Now().In(asiaTokyo)
	ret, err := app.client.Status(ctx, status, now.AddDate(0, 0, -7), now)
	if err != nil {
		return err
	}
	return json.NewEncoder(outStream).Encode(ret)
}
