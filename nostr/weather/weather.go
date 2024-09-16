package weather

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twpayne/go-meteomatics"
)

func GetWeather(countryCode string, zipCode string) []float64 {
	client := meteomatics.NewClient(
		meteomatics.WithBasicAuth(
			os.Getenv("METEOMATICS_USERNAME"),
			os.Getenv("METEOMATICS_PASSWORD"),
		),
	)

	cr, err := client.RequestCSV(
		context.Background(),
		meteomatics.TimeSlice{
			meteomatics.TimeNow,
			meteomatics.NowOffset(1 * time.Hour),
		},
		meteomatics.Parameter{
			Name:  meteomatics.ParameterTemperature,
			Level: meteomatics.LevelMeters(2),
			Units: meteomatics.UnitsFahrenheit,
		},
		meteomatics.Postal{
			CountryCode: countryCode,
			ZIPCode:     zipCode,
		},
		&meteomatics.RequestOptions{},
	)
	if err != nil {
		fmt.Println(err)
		return []float64{}
	}

	var values []float64
	for _, row := range cr.Rows {
		values = append(values, row.Values[0])
	}
	return values
}
