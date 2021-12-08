package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	geo "github.com/devincarr/lighthouse/geo"
	mqtt "github.com/devincarr/lighthouse/mqtt"
	quakes "github.com/devincarr/lighthouse/quakes"
	weather "github.com/devincarr/lighthouse/weather"

	dnsresolver "github.com/rs/dnscache"
)

type MqttConfig struct {
	Endpoint string
	Username string
	Password string
}

type EarthquakeConfig struct {
}

type WeatherConfig struct {
	Zones	[]string
	Email	string
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

var r = &dnsresolver.Resolver{}
var t = &http.Transport{
	DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := r.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			var dialer net.Dialer
			conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				break
			}
		}
		return
	},
}
var client = &http.Client{Transport: t}

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

	// Attempt to pre-cache dns responses
	for i := 0; i <= 5; i++ {
		addrs, err := r.LookupHost(context.Background(), "earthquake.usgs.gov")
		if err != nil {
			log.Println("error looking up: ", err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("Pre-cache DNS earthquake.usgs.gov: ", addrs)
		break
	}
	for i := 0; i <= 5; i++ {
		addrs, err := r.LookupHost(context.Background(), "api.weather.gov")
		if err != nil {
			log.Println("error looking up: ", err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("Pre-cache DNS api.weather.gov: ", addrs)
		break
	}

	// Call to refresh will refresh names in cache.
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for range t.C {
			r.Refresh(true)
		}
	}()

	local := geo.Point{c.Latitude, c.Longitude}
	for {
		alerts := make([]mqtt.Alert, 0)
		alerts, err := quakes.LocalQuakes(client, local, c.Distance)
		if err != nil {
			log.Printf("Unable to fetch local quake alerts: ", err)
		}
		fmt.Printf("Local earthquake alerts: %d\n", len(alerts))
		weatherAlerts, err := weather.LocalAlerts(client, c.Weather.Zones, c.Weather.Email)
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
