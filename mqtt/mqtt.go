package mqtt

import (
    "encoding/json"

    MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Alert struct {
    Type    string
    Id      string
    Message string
    Link    string
}

type Client struct {
    client  MQTT.Client
    sent    []string
    sentptr int32
}

// We are unlikely to have 15 alerts all sent recently enough to
// cause duplicate alert messages to not be captured in the sent array.
const MaxSentLength = 15

func NewClient(broker string, username string, password string) Client {
    opts := MQTT.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetUsername(username)
    opts.SetPassword(password)
    return Client{
        MQTT.NewClient(opts),
        make([]string, MaxSentLength),
        0,
    }
}

func (c *Client) Connect() error {
    if token := c.client.Connect(); token.Wait() && token.Error() != nil {
        return token.Error()
    }
    return nil
}

func (c *Client) Publish(alert Alert) error {
    for _, recent := range c.sent {
        if alert.Id == recent {
            // This alert has been sent recently already so we will skip it
            return nil
        }
    }
    b, err := json.Marshal(alert)
    if err != nil {
        return err
    }
    if token := c.client.Publish(alert.Type, 0, false, b); token.Wait() && token.Error() != nil {
        return token.Error()
    }
    // Add Alert to sent list
    c.sent[c.sentptr] = alert.Id
    if c.sentptr + 1 < MaxSentLength {
        c.sentptr += 1
    } else {
        c.sentptr = 0
    }
    return nil
}
