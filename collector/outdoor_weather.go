package collector

import (
	"fmt"

	"github.com/Tinkerforge/go-api-bindings/outdoor_weather_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterOutdoorWeatherBricklet(uid string) ([]Register, error) {
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
	for _ = range sids {
		cbID := d.RegisterSensorDataCallback(func(identifier uint8, temperature int16, humidity uint8) {
			idx := int(identifier) << 8
			b.Values <- Value{
				Index:    idx,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "Temperature of the air in °C",
				Name:     "temperature",
				Value:    float64(temperature) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 1,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
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
	}

	d.SetStationCallbackConfiguration(true)
	for _ = range stids {
		cbID := d.RegisterStationDataCallback(func(identifier uint8, temperature int16, humidity uint8, windSpeed uint32, gustSpeed uint32, rain uint32, windDirection uint8, batteryLow bool) {
			idx := int(identifier)<<8 + 65536
			b.Values <- Value{
				Index:    idx,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "Temperature of the air in °C",
				Name:     "temperature",
				Value:    float64(temperature) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 1,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "Humidity of the air in %rF",
				Name:     "humidity",
				Value:    float64(humidity),
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 2,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "WindSpeed in m/s",
				Name:     "wind_speed",
				Value:    float64(windSpeed) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 3,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "GustSpeed in m/s",
				Name:     "gust_speed",
				Value:    float64(gustSpeed) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 4,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "Rain in mm",
				Name:     "rain",
				Value:    float64(rain) / 10.0,
				Type:     prometheus.GaugeValue,
			}
			b.Values <- Value{
				Index:    idx + 5,
				DeviceID: outdoor_weather_bricklet.DeviceIdentifier,
				UID:      uid,
				SubID:    idx,
				Help:     "Wind Direction",
				Name:     "wind_direction",
				Value:    float64(windDirection),
				Type:     prometheus.GaugeValue,
			}
		})
		reg = append(reg, Register{
			Deregister: d.DeregisterStationDataCallback,
			ID:         cbID,
		})
	}
	return reg, nil
}
