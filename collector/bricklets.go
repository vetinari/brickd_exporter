package collector

import (
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
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

func (b *BrickdCollector) RegisterBarometerBricklet(uid string) []Register {
	d, _ := barometer_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	apID := d.RegisterAirPressureCallback(func(airPressure int32) {
		b.Values <- Value{
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
