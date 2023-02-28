package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// Collect is part of the prometheus.Collector interface
func (b *BrickdCollector) ExportMQTT(interval time.Duration) {
	for {
		b.exportMQTTOnce()
		time.Sleep(interval)
	}
}

type mqttData struct {
	Topic  string
	Labels map[string]string
	Data   map[string]interface{}
}

func (b *BrickdCollector) exportMQTTOnce() {
	b.RLock()
	defer b.RUnlock()

	data := map[string]interface{}{
		"connections_total": float64(b.ConnectCounter),
		"labels":            map[string]string{"brickd": b.Data.Address},
	}

	enc, err := json.Marshal(data)
	if err != nil {
		log.Errorf("failed to marshal json: %s", err)
		return
	}
	go b.MQTT.Client.Publish(b.MQTT.Topic.Name("brickd_exporter"), enc)

	mqData := make(map[string]mqttData)
	for _, vals := range b.Data.Values {
		for _, v := range vals {
			if v.UID == "" || b.ignored(v.UID) {
				continue
			}
			dev := fmt.Sprintf("%s.%d", v.UID, v.SensorID)
			if _, ok := mqData[dev]; !ok {
				md := mqttData{}
				labels := map[string]string{
					"uid":       v.UID,
					"brickd":    b.Data.Address,
					"id":        strconv.FormatInt(int64(v.DeviceID), 10),
					"type":      DeviceName(v.DeviceID),
					"sub_id":    strconv.Itoa(v.SensorID), // deprecated
					"sensor_id": strconv.Itoa(v.SensorID),
				}
				for k, v := range b.Labels {
					if _, exists := labels[k]; exists {
						continue
					}
					labels[k] = v
				}
				md.Topic = v.Name
				if sl, ok := b.SensorLabels[v.UID]; ok {
					if l, ok := sl[strconv.Itoa(v.SensorID)]; ok {
						for k, val := range l {
							if k == "mqtt_topic" {
								md.Topic = val
								continue
							}
							if _, exists := labels[k]; exists {
								continue
							}
							labels[k] = val
						}
					}
				}
				switch DeviceName(v.DeviceID) {
				case "Master Brick":
					md.Topic = "master_brick"
				case "HAT Brick":
					md.Topic = "hat_brick"
				case "HAT Zero Brick":
					md.Topic = "hat_zero_brick"
				}

				md.Labels = labels
				md.Data = make(map[string]interface{})
				mqData[dev] = md
			}
			mqData[dev].Data[v.Name] = v.Value
		}
	}
	for _, dev := range mqData {
		dev.Data["labels"] = dev.Labels
		enc, err := json.Marshal(dev.Data)
		if err != nil {
			log.Errorf("failed to marshal json: %s", err)
			return
		}
		go b.MQTT.Client.Publish(b.MQTT.Topic.Name(dev.Topic), enc)
	}
}
