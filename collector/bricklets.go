package collector

import (
	"fmt"

	"github.com/Tinkerforge/go-api-bindings/air_quality_bricklet"
	"github.com/Tinkerforge/go-api-bindings/ambient_light_v3_bricklet"
	"github.com/Tinkerforge/go-api-bindings/analog_in_v3_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/co2_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/uv_light_v2_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterAirQualityBricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
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
		b.Values <- Value{
			Index:    4,
			DeviceID: air_quality_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rH",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 100,
		}

	})
	if err := d.SetAllValuesCallbackConfiguration(b.CallbackPeriod, false); err != nil {
		return nil, fmt.Errorf("failed to set callback config for Air Quality Bricklet (uid=%s): %s", uid, err)
	}

	b.SetHAConfig("sensor", "aqi", "iaq_index", "", fmt.Sprintf("air_quality_bricklet%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "temperature", "temperature", "°C", fmt.Sprintf("air_quality_bricklet%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "atmospheric_pressure", "pressure", "hPa", fmt.Sprintf("air_quality_bricklet%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "humidity", "humidity", "%", fmt.Sprintf("air_quality_bricklet%s", uid), dev, 0, "")

	return []Register{
		{
			Deregister: d.DeregisterAllValuesCallback,
			ID:         cbID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterAnalogInV3Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
	d, err := analog_in_v3_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect AnalogInV3 Bricklet (uid=%s): %s", uid, err)
	}

	callbackID := d.RegisterVoltageCallback(func(voltage uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: analog_in_v3_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Voltage in V",
			Name:     "voltage",
			Type:     prometheus.GaugeValue,
			Value:    float64(voltage) / 1000.0,
		}
	})

	// set period to b.CallbackPeriod
	// valueHasToChange to false to also collect metrics if voltage is stable
	// Threshold is turned off and min/max zero to always collect metrics in fixed period
	d.SetVoltageCallbackConfiguration(b.CallbackPeriod, false, 'x', 0, 0)

	b.SetHAConfig("sensor", "voltage", "voltage", "V", fmt.Sprintf("analog_in_v3_bricklet_%s", uid), dev, 0, "")

	return []Register{
		{
			Deregister: d.DeregisterVoltageCallback,
			ID:         callbackID,
		},
	}, nil

}

func (b *BrickdCollector) RegisterHumidityBricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
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

	b.SetHAConfig("sensor", "humidity", "humidity", "%", fmt.Sprintf("humidity_bricklet_%s", uid), dev, 0, "")

	return []Register{
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         callbackID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterHumidityV2Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
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

	b.SetHAConfig("sensor", "humidity", "humidity", "%", fmt.Sprintf("humidity_bricklet_v2_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "temperature", "temperature", "°C", fmt.Sprintf("humidity_bricklet_v2_%s", uid), dev, 0, "")
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

func (b *BrickdCollector) RegisterBarometerBricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
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

	b.SetHAConfig("sensor", "atmospheric_pressure", "air_pressure", "hPa", fmt.Sprintf("barometer_bricklet_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "distance", "altitude", "m", fmt.Sprintf("barometer_bricklet_%s", uid), dev, 0, "")
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

func (b *BrickdCollector) RegisterBarometerV2Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
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
			Value:    float64(airPressure) / 1000.0,
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
			Value:    float64(altitude) / 1000.0,
		}
	})
	d.SetAltitudeCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	tempID := d.RegisterTemperatureCallback(func(temperature int32) {
		b.Values <- Value{
			Index:    2,
			DeviceID: barometer_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Temperature of the bricklet in °C",
			Name:     "bricklet_temperature",
			Type:     prometheus.GaugeValue,
			Value:    float64(temperature) / 100.0,
		}
	})
	d.SetTemperatureCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	b.SetHAConfig("sensor", "atmospheric_pressure", "air_pressure", "hPa", fmt.Sprintf("barometer_bricklet_v2_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "distance", "altitude", "m", fmt.Sprintf("barometer_bricklet_v2_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "temperature", "bricklet_temperature", "°C", fmt.Sprintf("barometer_bricklet_v2_%s", uid), dev, 0, "")

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

func (b *BrickdCollector) RegisterAmbientLightV3Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
	d, err := ambient_light_v3_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Ambient Light V3.0 (uid=%s): %s", uid, err)
	}

	ilID := d.RegisterIlluminanceCallback(func(illuminance uint32) {
		b.Values <- Value{
			Index:    0,
			DeviceID: ambient_light_v3_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Illuminance in Luxa",
			Name:     "illuminance",
			Type:     prometheus.GaugeValue,
			Value:    float64(illuminance) / 100,
		}
	})
	d.SetIlluminanceCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	b.SetHAConfig("sensor", "illuminance", "illuminance", "lx", fmt.Sprintf("ambient_light_v3_bricklet_%s", uid), dev, 0, "")
	return []Register{
		{
			Deregister: d.DeregisterIlluminanceCallback,
			ID:         ilID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterCO2V2Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
	d, err := co2_v2_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Barometer Bricklet V2.0 (uid=%s): %s", uid, err)
	}

	coID := d.RegisterCO2ConcentrationCallback(func(co2Concentration uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: co2_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "CO2 Concentration in PPM",
			Name:     "co2_concentration",
			Type:     prometheus.GaugeValue,
			Value:    float64(co2Concentration),
		}
	})
	d.SetCO2ConcentrationCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	huID := d.RegisterHumidityCallback(func(humidity uint16) {
		b.Values <- Value{
			Index:    1,
			DeviceID: co2_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rF",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 100,
		}
	})
	d.SetHumidityCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	tempID := d.RegisterTemperatureCallback(func(temperature int16) {
		b.Values <- Value{
			Index:    2,
			DeviceID: co2_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Temperature in °C",
			Name:     "temperature",
			Type:     prometheus.GaugeValue,
			Value:    float64(temperature) / 100,
		}
	})
	d.SetTemperatureCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	b.SetHAConfig("sensor", "carbon_dioxide", "co2_concentration", "ppm", fmt.Sprintf("co2_v2_bricklet_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "humidity", "humidity", "%", fmt.Sprintf("co2_v2_bricklet_%s", uid), dev, 0, "")
	b.SetHAConfig("sensor", "temperature", "temperature", "°C", fmt.Sprintf("co2_v2_bricklet_%s", uid), dev, 0, "")

	return []Register{
		{
			Deregister: d.DeregisterCO2ConcentrationCallback,
			ID:         coID,
		},
		{
			Deregister: d.DeregisterHumidityCallback,
			ID:         huID,
		},
		{
			Deregister: d.DeregisterTemperatureCallback,
			ID:         tempID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterUVLightV2Bricklet(dev *Device) ([]Register, error) {
	uid := dev.UID
	d, err := uv_light_v2_bricklet.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Ultra Violet Light V2.0 (uid=%s): %s", uid, err)
	}

	uvID := d.RegisterUVACallback(func(uva int32) {
		b.Values <- Value{
			Index:    0,
			DeviceID: uv_light_v2_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "UV in mW/m²",
			Name:     "uv",
			Type:     prometheus.GaugeValue,
			Value:    float64(uva) / 10,
		}
	})
	d.SetUVACallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	b.SetHAConfig("sensor", "", "uv", "mW/m²", fmt.Sprintf("uv_light_v2_bricklet_%s", uid), dev, 0, "")

	return []Register{
		{
			Deregister: d.DeregisterUVACallback,
			ID:         uvID,
		},
	}, nil
}
