package weather

import (
	"fmt"
	"io"
	"net/http"

	mqtt "github.com/devincarr/lighthouse/mqtt"

	geojson "github.com/devincarr/go.geojson"
)

const Topic = "alerts/weather"
const NWSAlertsUrl = "https://alerts.weather.gov/cap/wwaatmget.php?x=%s&y=1"
const UserAgentFormat = "(lighthouse, %s)"

func getNWSAlerts(client *http.Client, zone string, email string) (*geojson.FeatureCollection, error) {
	req, err := http.NewRequest("GET", "https://api.weather.gov/alerts/active/zone/"+zone, nil)
	useragent := fmt.Sprintf(UserAgentFormat, email)
	req.Header.Add("User-Agent", useragent)
	resp, err := client.Do(req)
	if err != nil {
		return new(geojson.FeatureCollection), err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return new(geojson.FeatureCollection), err
	}
	return geojson.UnmarshalFeatureCollection(body)
}

func LocalAlerts(client *http.Client, zones []string, email string) ([]mqtt.Alert, error) {
	alerts := make([]mqtt.Alert, 0)
	for _, zone := range zones {
		weatherAlerts, err := getNWSAlerts(client, zone, email)
		if err != nil {
			return alerts, err
		}
		for _, f := range weatherAlerts.Features {
			id := f.ID.(string)
			title, _ := f.PropertyString("headline")
			url := fmt.Sprintf(NWSAlertsUrl, zone)
			alerts = append(alerts, mqtt.Alert{Topic, id, title, url})
		}
	}
	return alerts, nil
}
