package collector

import (
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_v2_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterHumidityBricklet(uid string) []Register {
	d, _ := humidity_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	callbackID := d.RegisterHumidityCallback(func(humidity uint16) {
		b.Values <- Value{
			DeviceID: humidity_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rF",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 10.0,
		}
	})
	d.SetHumidityCallbackPeriod(10_000)
	return []Register{
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         callbackID,
		},
	}
}

func (b *BrickdCollector) RegisterHumidityV2Bricklet(uid string) []Register {
	d, _ := humidity_v2_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	humID := d.RegisterHumidityCallback(func(humidity uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: humidity_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rF",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 100.0,
		}
	})
	d.SetHumidityCallbackConfiguration(10_000, true, 'x', 0, 0)

	tempID := d.RegisterTemperatureCallback(func(temperature int16) {
		b.Values <- Value{
			Index:    1,
			DeviceID: humidity_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Temperature of the air in °C",
			Name:     "temperature",
			Type:     prometheus.GaugeValue,
			Value:    float64(temperature) / 100.0,
		}
	})
	d.SetTemperatureCallbackConfiguration(10_000, true, 'x', 0, 0)

	return []Register{
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         humID,
		},
		{
			Deregister: d.DeregisterTemperatureCallback,
			ID:         tempID,
		},
	}
}

func (b *BrickdCollector) RegisterBarometerBricklet(uid string) []Register {
	d, _ := barometer_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	apID := d.RegisterAirPressureCallback(func(airPressure int32) {
		b.Values <- Value{
			Index:    0,
			DeviceID: barometer_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Air Pressure in hPa",
			Name:     "air_pressure",
			Type:     prometheus.GaugeValue,
			Value:    float64(airPressure) * 1000.0,
		}
	})
	d.SetAirPressureCallbackPeriod(10_000)

	altID := d.RegisterAltitudeCallback(func(altitude int32) {
		b.Values <- Value{
			Index:    1,
			DeviceID: barometer_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Altitude in m",
			Name:     "altitude",
			Type:     prometheus.GaugeValue,
			Value:    float64(altitude) * 100.0,
		}
	})
	d.SetAltitudeCallbackPeriod(10_000)

	return []Register{
		{
			Deregister: d.DeregisterAirPressureCallback,
			ID:         apID,
		},
		{
			Deregister: d.DeregisterAltitudeCallback,
			ID:         altID,
		},
	}
}

func (b *BrickdCollector) RegisterBarometerV2Bricklet(uid string) []Register {
	d, _ := barometer_v2_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	apID := d.RegisterAirPressureCallback(func(airPressure int32) {
		b.Values <- Value{
			Index:    0,
			DeviceID: barometer_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Air Pressure in hPa",
			Name:     "air_pressure",
			Type:     prometheus.GaugeValue,
			Value:    float64(airPressure) * 1000.0,
		}
	})
	d.SetAirPressureCallbackConfiguration(10_000, true, 'x', 0, 0)

	altID := d.RegisterAltitudeCallback(func(altitude int32) {
		b.Values <- Value{
			Index:    1,
			DeviceID: barometer_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Altitude in m",
			Name:     "altitude",
			Type:     prometheus.GaugeValue,
			Value:    float64(altitude) * 1000.0,
		}
	})
	d.SetAltitudeCallbackConfiguration(10_000, true, 'x', 0, 0)

	tempID := d.RegisterTemperatureCallback(func(temperature int32) {
		b.Values <- Value{
			Index:    2,
			DeviceID: barometer_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Temperature in °C",
			Name:     "temperature",
			Type:     prometheus.GaugeValue,
			Value:    float64(temperature) * 100.0,
		}
	})
	d.SetTemperatureCallbackConfiguration(10_000, true, 'x', 0, 0)

	return []Register{
		{
			Deregister: d.DeregisterAirPressureCallback,
			ID:         apID,
		},
		{
			Deregister: d.DeregisterAltitudeCallback,
			ID:         altID,
		},
		{
			Deregister: d.DeregisterTemperatureCallback,
			ID:         tempID,
		},
	}
}
