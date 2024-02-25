package collector

import "strings"

func (b *BrickdCollector) PublishHAConfig(dev *Device) {
	if dev.UnitOfMeasurement == "" {
		return
	}
	topic = b.MQTT.HomeAssistant.DiscoveryTopic
	if topic == "" {
		topic = "homeassistant/"
	}
	if !strings.HasSuffix(baseTopic, "/") {
		topic += "/"
	}
	topic += "sensor/brickd_" + dev.UID + "/config"

	cfg := &HAConfig{
		UniqueID:          "brickd_" + dev.UID,
		StateTopic:        "brickd/" + b.SensorTopic(dev.UID),
		UnitOfMeasurement: dev.UnitOfMeasurement,
	}
	switch dev.DeviceID {
	case 288: // Outdoor Weather Bricklet
	}
}

type HAConfig struct {
	DeviceClass       string   `json:"device_class"`
	StateTopic        string   `json:"state_topic"`
	UnitOfMeasurement string   `json:"unit_of_measurement"`
	ValueTemplate     string   `json:"value_template"`
	UniqueID          string   `json:"unique_id"`
	Device            HADevice `json:"device"`
}

type HADevice struct {
	Identifiers      []string `json:"identifiers"`
	Name             string   `json:"name"`
	Manufacturer     string   `json:"manufacturer"`
	Model            string   `json:"model"`
	SerialNumber     string   `json:"serial_number,omitempty"`
	HWVersion        string   `json:"hw_version,omitempty"`
	SWVersion        string   `json:"sw_version,omitempty"`
	ConfigurationURL string   `json:"configuration_url,omitempty"`
}
