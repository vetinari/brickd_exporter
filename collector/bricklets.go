package collector

import (
	"fmt"

	"github.com/Tinkerforge/go-api-bindings/air_quality_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_v2_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterAirQualityBricklet(uid string) ([]Register, error) {
	d, err := air_quality_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Air Quality Bricklet (uid=%s): %s", uid, err)
	}

	cbID := d.RegisterAllValuesCallback(func(iaqIndex int32, iaqIndexAccuracy uint8, temperature int32, humidity int32, airPressure int32) {
		b.Values <- Value{
			Index:    0,
			DeviceID: air_quality_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "IAQ Index Value",
			Name:     "iaq_index",
			Type:     prometheus.GaugeValue,
			Value:    float64(iaqIndex),
		}
		b.Values <- Value{
			Index:    1,
			DeviceID: air_quality_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "IAQ Index Accuracy",
			Name:     "iaq_index_accuracy",
			Type:     prometheus.GaugeValue,
			Value:    float64(iaqIndexAccuracy),
		}
		b.Values <- Value{
			Index:    2,
			DeviceID: air_quality_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Temperature of the air in °C",
			Name:     "temperature",
			Type:     prometheus.GaugeValue,
			Value:    float64(temperature) / 100,
		}
		b.Values <- Value{
			Index:    3,
			DeviceID: air_quality_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Air Pressure in hPa",
			Name:     "pressure",
			Type:     prometheus.GaugeValue,
			Value:    float64(airPressure) / 100,
		}

	})
	if err := d.SetAllValuesCallbackConfiguration(b.CallbackPeriod, false); err != nil {
		return nil, fmt.Errorf("failed to set callback config for Air Quality Bricklet (uid=%s): %s", uid, err)
	}
	return []Register{
		{
			Deregister: d.DeregisterAllValuesCallback,
			ID:         cbID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterHumidityBricklet(uid string) ([]Register, error) {
	d, err := humidity_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Humidity Bricklet (uid=%s): %s", uid, err)
	}

	callbackID := d.RegisterHumidityCallback(func(humidity uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: humidity_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rF",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 10.0,
		}
	})
	d.SetHumidityCallbackPeriod(b.CallbackPeriod)
	return []Register{
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         callbackID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterHumidityV2Bricklet(uid string) ([]Register, error) {
	d, err := humidity_v2_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Humidity Bricklet V2.0 (uid=%s): %s", uid, err)
	}

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
	d.SetHumidityCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

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
	d.SetTemperatureCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	return []Register{
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         humID,
		},
		{
			Deregister: d.DeregisterTemperatureCallback,
			ID:         tempID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterBarometerBricklet(uid string) ([]Register, error) {
	d, err := barometer_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Barometer Bricklet (uid=%s): %s", uid, err)
	}

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
	d.SetAirPressureCallbackPeriod(b.CallbackPeriod)

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
	d.SetAltitudeCallbackPeriod(b.CallbackPeriod)

	return []Register{
		{
			Deregister: d.DeregisterAirPressureCallback,
			ID:         apID,
		},
		{
			Deregister: d.DeregisterAltitudeCallback,
			ID:         altID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterBarometerV2Bricklet(uid string) ([]Register, error) {
	d, err := barometer_v2_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Barometer Bricklet V2.0 (uid=%s): %s", uid, err)
	}

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
	d.SetAirPressureCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

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
	d.SetAltitudeCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

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
	d.SetTemperatureCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

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
	}, nil
}
