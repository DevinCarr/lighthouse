package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	geo "github.com/devincarr/lighthouse/geo"
	mqtt "github.com/devincarr/lighthouse/mqtt"
	quakes "github.com/devincarr/lighthouse/quakes"
	weather "github.com/devincarr/lighthouse/weather"
)

type MqttConfig struct {
	Endpoint string
	Username string
	Password string
}

type EarthquakeConfig struct {
}

type WeatherConfig struct {
	Zone  string
	Email string
}

type Config struct {
	Latitude   float64
	Longitude  float64
	Distance   float64
	Interval   string
	Mqtt       MqttConfig
	Earthquake EarthquakeConfig
	Weather    WeatherConfig
}

func main() {
	// Parse config
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "Config file")
	flag.Parse()
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Unable to open config file: ", err)
		return
	}
	var c Config
	if err := json.Unmarshal(file, &c); err != nil {
		log.Fatal("Unable to parse config file: ", err)
		return
	}

	interval, err := time.ParseDuration(c.Interval)
	if err != nil {
		log.Fatal("Unable to parse interval Duration: ", err)
		return
	}

	mqttClient := mqtt.NewClient(c.Mqtt.Endpoint, c.Mqtt.Username, c.Mqtt.Password)
	if err := mqttClient.Connect(); err != nil {
		panic(err)
	}

	local := geo.Point{c.Latitude, c.Longitude}
	for {
		alerts := make([]mqtt.Alert, 0)
		alerts, err := quakes.LocalQuakes(local, c.Distance)
		if err != nil {
			log.Printf("Unable to fetch local quake alerts: ", err)
		}
		fmt.Printf("Local earthquake alerts: %d\n", len(alerts))
		weatherAlerts, err := weather.LocalAlerts(c.Weather.Zone, c.Weather.Email)
		if err != nil {
			log.Printf("Unable to fetch local weather alerts: ", err)
		}
		fmt.Printf("Local weather alerts: %d\n", len(weatherAlerts))
		alerts = append(alerts, weatherAlerts...)
		for _, alert := range alerts {
			if err := mqttClient.Publish(alert); err != nil {
				log.Fatal("Unable to publish alert: ", err)
			}
		}
		time.Sleep(interval)
	}
}
