package quakes

import (
	"io"
	"net/http"

	geo "github.com/devincarr/lighthouse/geo"
	mqtt "github.com/devincarr/lighthouse/mqtt"

	geojson "github.com/devincarr/go.geojson"
)

const Topic = "alerts/earthquakes"

func getQuakes() (*geojson.FeatureCollection, error) {
	resp, err := http.Get("https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson")
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

func checkQuakes(fc *geojson.FeatureCollection, local geo.Point, dist float64) []*geojson.Feature {
	localQuakes := make([]*geojson.Feature, 0)
	for _, f := range fc.Features {
		quakeLat := f.Geometry.Point[1]
		quakeLong := f.Geometry.Point[0]

		quakePoint := geo.Point{quakeLat, quakeLong}
		hdist := geo.Haversine(quakePoint, local)
		if hdist <= dist {
			localQuakes = append(localQuakes, f)
		}
	}
	return localQuakes
}

func LocalQuakes(local geo.Point, distance float64) ([]mqtt.Alert, error) {
	alerts := make([]mqtt.Alert, 0)
	quakes, err := getQuakes()
	if err != nil {
		return alerts, err
	}
	localQuakes := checkQuakes(quakes, local, distance)
	for _, f := range localQuakes {
		id := f.ID.(string)
		place, _ := f.PropertyString("place")
		url, _ := f.PropertyString("url")
		alerts = append(alerts, mqtt.Alert{Topic, id, place, url})
	}
	return alerts, nil
}
