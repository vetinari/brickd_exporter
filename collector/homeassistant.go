package collector

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (b *BrickdCollector) SetHAConfig(typ, devClass, valueName, unit, uniqueID string, dev *Device, idx int) {
	if b.MQTT == nil || !b.MQTT.Enabled || !b.MQTT.HomeAssistant.Enabled {
		return
	}
	topic := b.MQTT.HomeAssistant.DiscoveryBase
	if topic == "" {
		topic = "homeassistant/"
	}
	topic += typ + "/brickd_" + uniqueID + "_" + valueName + "/config"
	id := b.DefaultTopic(dev)

	cfg := &HAConfig{
		DeviceClass:       devClass,
		UniqueID:          "brickd_" + uniqueID + "_" + valueName,
		ObjectID:          "brickd_" + uniqueID + "_" + valueName,
		Name:              "brickd_" + uniqueID + "_" + valueName,
		StateTopic:        "brickd/" + b.SensorTopic(dev, idx),
		UnitOfMeasurement: unit,
		ValueTemplate:     fmt.Sprintf("{{ value_json.%s }}", valueName),
		Device: HADevice{
			Name:         "Brickd: " + b.Address + " / " + DeviceName(dev.DeviceID),
			Identifiers:  []string{id},
			HWVersion:    dev.HardwareVersion,
			SWVersion:    dev.FirmwareVersion,
			Manufacturer: "Tinkerforge GmbH",
			Model:        DeviceName(dev.DeviceID),
		},
		Origin: HAOrigin{
			Name:       "brickd",
			SWVersion:  Version,
			SupportURL: "https://github.com/vetinari/brickd_exporter",
		},
	}
	enc, err := json.Marshal(cfg)
	if err != nil {
		log.Errorf("failed to marshal HA Config: %s", err)
		return
	}
	log.Infof("publishing HA config to %s: %s", topic, string(enc))
	go b.MQTT.Client.Publish(topic, enc)
}

type HAConfig struct {
	Name              string   `json:"name"`
	DeviceClass       string   `json:"device_class"`
	StateTopic        string   `json:"state_topic"`
	UnitOfMeasurement string   `json:"unit_of_measurement"`
	ValueTemplate     string   `json:"value_template"`
	UniqueID          string   `json:"unique_id"`
	ObjectID          string   `json:"object_id"`
	Device            HADevice `json:"device"`
	Origin            HAOrigin `json:"origin"`
}

type HAOrigin struct {
	Name       string `json:"name"`
	SWVersion  string `json:"sw_version"`
	SupportURL string `json:"support_url"`
}

type HADevice struct {
	Identifiers      []string `json:"identifiers"`
	Name             string   `json:"name,omitempty"`
	Manufacturer     string   `json:"manufacturer"`
	Model            string   `json:"model"`
	SerialNumber     string   `json:"serial_number,omitempty"`
	HWVersion        string   `json:"hw_version,omitempty"`
	SWVersion        string   `json:"sw_version,omitempty"`
	ConfigurationURL string   `json:"configuration_url,omitempty"`
}
