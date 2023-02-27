package mqtt

import (
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type MQTT struct {
	Enabled bool    `yaml:"enabled"`
	Broker  *Broker `yaml:"broker"`
	Topic   Topic   `yaml:"topic"`
	Client  *Client `yaml:"-"`
}

type Topic string

func (t *Topic) Name(name string) string {
	if !strings.HasSuffix(string(*t), "/") {
		return string(*t) + "/" + name
	}
	return string(*t) + name
}

type Broker struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	ClientID string `yaml:"client_id"`
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.WithFields(log.Fields{
		"type":    "mqtt",
		"topic":   msg.Topic(),
		"payload": msg.Payload(),
	}).Debug("received message")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	r := client.OptionsReader()
	s := r.Servers()
	log.WithFields(log.Fields{
		"type": "mqtt",
		"urls": fmt.Sprintf("%+v", s),
	}).Infof("connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	r := client.OptionsReader()
	s := r.Servers()
	log.WithFields(log.Fields{
		"type":  "mqtt",
		"error": err.Error(),
		"urls":  fmt.Sprintf("%+v", s),
	}).Warn("connection lost")
}

type Client struct {
	c mqtt.Client
}

func NewClient(broker *Broker) (*Client, error) {
	if broker.Port == 0 {
		broker.Port = 1833
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker.Host, broker.Port))
	opts.SetClientID(broker.ClientID)
	opts.SetUsername(broker.Username)
	opts.SetPassword(broker.Password)

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to broker: %w", token.Error())
	}
	return &Client{c: client}, nil
}

func (c *Client) Publish(topic string, data []byte) {
	token := c.c.Publish(topic, 1, false, data)
	token.Wait()
}

func (c *Client) Client() mqtt.Client {
	return c.c
}
