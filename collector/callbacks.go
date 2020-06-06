package collector

import (
	"fmt"
	"github.com/Tinkerforge/go-api-bindings/ipconnection"
	log "github.com/sirupsen/logrus"
)

// OnConnect is called when the brickd collector (re-)connects to the brickd
func (b *BrickdCollector) OnConnect(reason ipconnection.DisconnectReason) {
	var why string
	switch reason {
	case ipconnection.ConnectReasonRequest:
		why = "Connection established after request from user."
	case ipconnection.ConnectReasonAutoReconnect:
		why = "Connection after auto-reconnect."
	}
	log.Infof("connected to brickd: %s", why)

	// Authenticate first...
	if b.Password != "" {
		err := b.Connection.Authenticate(b.Password)
		if err != nil {
			log.Errorf("Could not authenticate:", err)
			return
		}
		log.Debugf("Authentication succeded")
	}
	b.Connection.SetAutoReconnect(true)
	b.Connection.Enumerate() // call now, so we get devices when we initially connect
}

// OnDisconnect is called when the brickd collector disconnects from the brickd
func (b *BrickdCollector) OnDisconnect(reason ipconnection.DisconnectReason) {
	var why string
	switch reason {
	case ipconnection.DisconnectReasonRequest:
		why = "Disconnect was requested by user."
	case ipconnection.DisconnectReasonError:
		why = "Disconnect because of an unresolvable error."
	case ipconnection.DisconnectReasonShutdown:
		why = "Disconnect initiated by Brick Daemon or WIFI/Ethernet Extension."
	}
	log.Info("disconnected from brickd: %s", why)

	b.Lock()
	for _, dev := range b.Data.Devices {
		if reg, ok := b.Registry[dev.UID]; ok {
			for _, d := range reg {
				log.Debugf("deregistering callback %d of %s", d.ID, dev.UID)
				d.Deregister(d.ID)
			}
			delete(b.Registry, dev.UID)
			delete(b.Data.Values, dev.UID)
		}
	}
	b.Data.Devices = make(map[string]*Device)
	b.Unlock()
}

// OnEnumerate receives the callbacks from the Enumerate() call
func (b *BrickdCollector) OnEnumerate(
	uid string,
	connectedUid string,
	position rune,
	hardwareVersion [3]uint8,
	firmwareVersion [3]uint8,
	deviceIdentifier uint16,
	enumerationType ipconnection.EnumerationType) {

	dev := &Device{
		UID:             uid,
		ConnectedUID:    connectedUid,
		Position:        position,
		HardwareVersion: fmt.Sprintf("%d.%d.%d", hardwareVersion[0], hardwareVersion[1], hardwareVersion[2]),
		FirmwareVersion: fmt.Sprintf("%d.%d.%d", firmwareVersion[0], firmwareVersion[1], firmwareVersion[2]),
		DeviceID:        deviceIdentifier,
	}
	if enumerationType == ipconnection.EnumerationTypeAvailable {
		dev.Available = true
	}
	log.Debugf("got device %#+v", dev)
	b.Lock()
	defer b.Unlock()

	regFunc, ok := b.Devices[dev.DeviceID]
	if !ok {
		log.Debugf("not setting callback for: %s %d", dev.UID, dev.DeviceID)
		return
	}

	if _, ok := b.Registry[dev.UID]; ok {
		log.Debugf("callback already registered: %s %d", dev.UID, dev.DeviceID)
		for _, reg := range b.Registry[dev.UID] {
			log.Debugf("callback for %s %d -> %d", dev.UID, dev.DeviceID, reg.ID)
		}
		return
	}

	reg := regFunc(dev.UID)
	if len(reg) == 0 {
		log.Debugf("no registry returned from %s %d", dev.UID, dev.DeviceID)
		return
	}
	b.Data.Devices[dev.UID] = dev
	b.Registry[dev.UID] = reg
	for _, reg := range b.Registry[dev.UID] {
		log.Debugf("callback registered for %s %d -> %d", dev.UID, dev.DeviceID, reg.ID)
	}
}
