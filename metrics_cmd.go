package healthplanet

import (
	"context"
	"fmt"
	"io"
	"time"
)

var asiaTokyo *time.Location

func init() {
	var err error
	asiaTokyo, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

func metricsCmd(ctx context.Context, argv []string, outStream, errWtream io.Writer) error {
	app := getApp(ctx)
	now := time.Now()
	var metrics = map[string]*Data{}
	var height string // for BMI
	for _, status := range []string{"innerscan", "sphygmomanometer"} {
		ret, err := app.client.Status(ctx, status, now.AddDate(0, 0, -7), now)
		if err != nil {
			return err
		}
		if height == "" {
			height = ret.Height
		}
		for _, d := range ret.Data {
			prev, ok := metrics[d.Tag]
			if !ok || d.Date > prev.Date {
				metrics[d.Tag] = d
			}
		}
	}

	var metricsKeys = [][2]string{
		{"6021", "body.weight"},
		{"6022", "body.fat_rate"},
		{"622E", "blood_pressure.systolic"},
		{"622F", "blood_pressure.diastolic"},
	}
	for _, keys := range metricsKeys {
		tag, key := keys[0], keys[1]
		d, ok := metrics[tag]
		if !ok {
			continue
		}
		ti, err := time.ParseInLocation(dataTimeLayout, d.Date, asiaTokyo)
		if err != nil {
			return err
		}
		fmt.Fprintf(outStream, "%s\t%s\t%d\n", key, d.KeyData, ti.Unix())
	}
	return nil
}
