package collector

import (
	"fmt"
	"time"

	"github.com/Tinkerforge/go-api-bindings/hat_brick"
	"github.com/Tinkerforge/go-api-bindings/hat_zero_brick"
	"github.com/Tinkerforge/go-api-bindings/ipconnection"
	"github.com/Tinkerforge/go-api-bindings/master_brick"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func (b *BrickdCollector) RegisterMasterBrick(uid string) ([]Register, error) {
	m, err := master_brick.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Master Brick (uid=%s): %s", uid, err)
	}

	hasEthernet, err := m.IsEthernetPresent()
	if err != nil {
		hasEthernet = false
	}
	if hasEthernet {
		log.Debugf("ethernet extension is present")
		go b.PollEthernetState(m, uid)
	}

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
	m.SetStackCurrentCallbackPeriod(b.CallbackPeriod)

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
	m.SetStackVoltageCallbackPeriod(b.CallbackPeriod)

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
	m.SetUSBVoltageCallbackPeriod(b.CallbackPeriod)

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
		{
			Deregister: b.CloseEthernetState,
			ID:         EthernetCallbackID,
		},
	}, nil
}

func (b *BrickdCollector) CloseEthernetState(_ uint64) {
	close(b.EthernetState)
}

func (b *BrickdCollector) PollEthernetState(m master_brick.MasterBrick, uid string) {
	b.EthernetState = make(chan interface{})
	go func() {
		select {
		case <-b.EthernetState:
			return
		default:
			for {
				if b.Connection.GetConnectionState() != ipconnection.ConnectionStateConnected {
					time.Sleep(time.Duration(b.CallbackPeriod) * time.Millisecond)
					continue
				}
				_, _, _, _, rxCount, txCount, _, err := m.GetEthernetStatus()
				log.Debugf("ethernet connected: rx %d / tx %d", rxCount, txCount)
				if err != nil {
					log.Infof("failed to get ethernet status: %s", err)
					time.Sleep(time.Duration(b.CallbackPeriod) * time.Millisecond)
					continue
				}

				b.Values <- Value{
					Index:    3,
					DeviceID: master_brick.DeviceIdentifier,
					UID:      uid,
					Help:     "Received bytes by Ethernet Extension",
					Name:     "ethernet_received",
					Type:     prometheus.CounterValue,
					Value:    float64(rxCount),
				}
				b.Values <- Value{
					Index:    4,
					DeviceID: master_brick.DeviceIdentifier,
					UID:      uid,
					Help:     "Transmitted bytes by Ethernet Extension",
					Name:     "ethernet_transmitted",
					Type:     prometheus.CounterValue,
					Value:    float64(txCount),
				}
				time.Sleep(time.Duration(b.CallbackPeriod) * time.Millisecond)
			}
		}
	}()
}

func (b *BrickdCollector) RegisterZeroHatBrick(uid string) ([]Register, error) {
	h, err := hat_zero_brick.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Zero Hat Brick (uid=%s): %s", uid, err)
	}

	vID := h.RegisterUSBVoltageCallback(func(current uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: hat_zero_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "Voltage of the Zero Hat in V",
			Name:     "voltage",
			Type:     prometheus.GaugeValue,
			Value:    float64(current) / 1000.0,
		}
	})
	h.SetUSBVoltageCallbackConfiguration(b.CallbackPeriod, true, 'x', 0, 0)

	return []Register{
		{
			Deregister: h.DeregisterUSBVoltageCallback,
			ID:         vID,
		},
	}, nil
}

func (b *BrickdCollector) RegisterHatBrick(uid string) ([]Register, error) {
	h, err := hat_brick.New(uid, &b.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Hat Brick (uid=%s): %s", uid, err)
	}

	callbackID := h.RegisterVoltagesCallback(func(voltageUSB uint16, voltageDC uint16) {
		b.Values <- Value{
			Index:    0,
			DeviceID: hat_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "Voltage of the Hat USB in V",
			Name:     "voltage_usb",
			Type:     prometheus.GaugeValue,
			Value:    float64(voltageUSB) / 1000.0,
		}
		b.Values <- Value{
			Index:    1,
			DeviceID: hat_brick.DeviceIdentifier,
			UID:      uid,
			Help:     "Voltage of the Hat DC in V",
			Name:     "voltage_dc",
			Type:     prometheus.GaugeValue,
			Value:    float64(voltageDC) / 1000.0,
		}
	})
	h.SetVoltagesCallbackConfiguration(b.CallbackPeriod, false)

	return []Register{
		{
			Deregister: h.DeregisterVoltagesCallback,
			ID:         callbackID,
		},
	}, nil
}
