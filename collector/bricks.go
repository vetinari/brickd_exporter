package collector

import (
	// log "github.com/sirupsen/logrus"
	"github.com/Tinkerforge/go-api-bindings/master_brick"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *BrickdCollector) RegisterMasterBrick(uid string) []Register {
	m, _ := master_brick.New(uid, &b.Connection)
	// FIXME handle error here and return nil
	currID := m.RegisterStackCurrentCallback(func(current uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: master_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "Current of the stack in A",
			Name:     "stack_current",
			Type:     prometheus.GaugeValue,
			Value:    float64(current) / 1000.0,
		}
	})
	m.SetStackCurrentCallbackPeriod(10_000)

	voltID := m.RegisterStackVoltageCallback(func(voltage uint16) {
		b.Values <- Value{
			Index:    1,
			DeviceID: master_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "Voltage of the stack in V",
			Name:     "stack_voltage",
			Type:     prometheus.GaugeValue,
			Value:    float64(voltage) / 1000.0,
		}
	})
	m.SetStackVoltageCallbackPeriod(10_000)

	usbVID := m.RegisterUSBVoltageCallback(func(voltage uint16) {
		b.Values <- Value{
			Index:    2,
			DeviceID: master_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "USB Voltage of the stack in V",
			Name:     "usb_voltage",
			Type:     prometheus.GaugeValue,
			Value:    float64(voltage) / 1000.0,
		}
	})
	m.SetUSBVoltageCallbackPeriod(10_000)

	return []Register{
		{
			Deregister: m.DeregisterStackCurrentCallback,
			ID:         currID,
		},
		{
			Deregister: m.DeregisterStackVoltageCallback,
			ID:         voltID,
		},
		{
			Deregister: m.DeregisterUSBVoltageCallback,
			ID:         usbVID,
		},
	}
}
