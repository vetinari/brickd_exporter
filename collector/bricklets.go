package collector

import (
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterHumidityBricklet(uid string) []Register {
	h, _ := humidity_bricklet.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	callbackID := h.RegisterHumidityCallback(func(humidity uint16) {
		b.Values <- Value{
			DeviceID: humidity_bricklet.DeviceIdentifier,
			UID:      uid,
			Help:     "Humidity of the air in %rF",
			Name:     "humidity",
			Type:     prometheus.GaugeValue,
			Value:    float64(humidity) / 10.0,
		}
	})
	h.SetHumidityCallbackPeriod(10_000)
	return []Register{
		{
			Deregister: h.DeregisterHumidityCallback,
			ID:         callbackID,
		},
	}
}
