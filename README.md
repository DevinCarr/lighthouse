# Lighthouse

Notification system for local earthquakes from USGS and weather alerts from NWS.

## Usage

`$ ./lighthouse -config config.json`

### Config

```json
{
    "Latitude": 61.181,
    "Longitude": -149.989,
    "Distance": 500.0, // in km
    "Interval": "1m", // interval to check sources
    "Weather": {
        "Zone": "AKZ101", // NWS weather zone (https://alerts.weather.gov/index.php)
        "Email": "email@email.com" // Email for NWS API
    },
    "Mqtt": {
        "Endpoint": "tcp://homeassistant.local:1883",
        "Username": "mqtt-client",
        "Password": "fancy-password-here"
    }
}
```

### MQTT/HomeAssistant Integration

For any alert published by the monitored systems, the message will be packed into a MQTT payload for easy parsing in HomeAssistant Automation.

Example MQTT payload (earthquake):

```json
{
    "Type": "alerts/earthquakes",
    "Id": "nc73648785",
    "Message": "M 2.1 - 7km NE of San Martin, CA",
    "Link":"https://earthquake.usgs.gov/earthquakes/eventpage/nc73648785"
}
```

Example Home Assistant Automation (earthquake):

```yaml
alias: Send Push Notification for Local Earthquake Alerts
description: 'Send Push Notification for Local Earthquake Alerts'
trigger:
  - platform: mqtt
    topic: alerts/earthquakes
condition: []
action:
  - service: notify.mobile_app_<device>
    data:
      title: Earthquake Alert
      message: '{{ trigger.payload_json.Message }}'
      data:
        actions:
          - action: URI
            title: Open USGS Alert
            uri: '{{ trigger.payload_json.Link }}'
mode: parallel
max: 5
```
