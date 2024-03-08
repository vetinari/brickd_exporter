package collector

import (
	"fmt"
	"strconv"

	"github.com/Tinkerforge/go-api-bindings/outdoor_weather_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterOutdoorWeatherBricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
	d, err := outdoor_weather_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Outdoor Weather Bricklet V2.0 (uid=%s): %s", uid, err)
	}

	sids, err := d.GetSensorIdentifiers()
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor identifiers: %s", err)
	}

	stids, err := d.GetStationIdentifiers()
	if err != nil {
		return nil, fmt.Errorf("failed to get station identifiers: %s", err)
	}
	var reg []Register

	d.SetSensorCallbackConfiguration(true)
	for _, sid := range sids {
		cbID := d.RegisterSensorDataCallback(func(identifier uint8, temperature int16, humidity uint8) {
			idx := int(identifier) << 8
			b.Values <- Value{
				Index:    idx,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Temperature of the air in °C",
				Name:     "temperature",
				Value:    float64(temperature) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 1,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Humidity of the air in %rF",
				Name:     "humidity",
				Value:    float64(humidity),
				Type:     prometheus.GaugeValue,
			}
		})
		reg = append(reg, Register{
			Deregister: d.DeregisterSensorDataCallback,
			ID:         cbID,
		})
		idx := int(sid) << 8
		uniqueID := fmt.Sprintf("outdoor_weather_bricklet_%s_%d", uid, idx)
		b.SetHAConfig("sensor", "temperature", "temperature", "°C", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("sensor", "humidity", "humidity", "%", uniqueID, dev, idx, strconv.Itoa(idx))
	}

	d.SetStationCallbackConfiguration(true)
	for _, stid := range stids {
		cbID := d.RegisterStationDataCallback(func(identifier uint8, temperature int16, humidity uint8, windSpeed uint32, gustSpeed uint32, rain uint32, windDirection uint8, batteryLow bool) {
			idx := int(identifier)<<8 + 65536
			b.Values <- Value{
				Index:    idx,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Temperature of the air in °C",
				Name:     "temperature",
				Value:    float64(temperature) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 1,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Humidity of the air in %rF",
				Name:     "humidity",
				Value:    float64(humidity),
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 2,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "WindSpeed in m/s",
				Name:     "wind_speed",
				Value:    float64(windSpeed) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 3,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "GustSpeed in m/s",
				Name:     "gust_speed",
				Value:    float64(gustSpeed) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 4,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Rain in mm",
				Name:     "rain",
				Value:    float64(rain) / 10.0,
				Type:     prometheus.CounterValue,
			}
			b.Values <- Value{
				Index:    idx + 5,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Wind Direction",
				Name:     "wind_direction",
				Value:    float64(windDirection),
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 6,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SensorID: idx,
				Help:     "Battery Status Low",
				Name:     "battery_low",
				Value:    bool2Float(batteryLow),
				Type:     prometheus.GaugeValue,
			}
		})
		reg = append(reg, Register{
			Deregister: d.DeregisterStationDataCallback,
			ID:         cbID,
		})
		idx := int(stid)<<8 + 65536
		uniqueID := fmt.Sprintf("outdoor_weather_bricklet_%s_%d", uid, idx)

		b.SetHAConfig("sensor", "temperature", "temperature", "°C", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("sensor", "humidity", "humidity", "%", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("sensor", "wind_speed", "wind_speed", "m/s", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("sensor", "wind_speed", "gust_speed", "m/s", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("sensor", "precipitation", "rain", "mm", uniqueID, dev, idx, strconv.Itoa(idx))
		b.SetHAConfig("binary_sensor", "battery", "battery_low", "", uniqueID, dev, idx, strconv.Itoa(idx))
	}
	return reg, nil
}

func bool2Float(v bool) float64 {
	if v {
		return 1.0
	}
	return 0.0
}
