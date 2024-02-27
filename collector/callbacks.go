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

	b.Lock()
	b.ConnectCounter += 1
	b.Unlock()

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
	log.Infof("disconnected from brickd: %s", why)

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

	if b.ignored(uid) {
		return
	}
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
		log.Debugf("no callbacks available for %s (uid=%s)", DeviceName(dev.DeviceID), dev.UID)
		return
	}

	if _, ok := b.Registry[dev.UID]; ok {
		log.Debugf("callback already registered for %s (uid=%s)", DeviceName(dev.DeviceID), dev.UID)
		for _, reg := range b.Registry[dev.UID] {
			log.Debugf("callback for %s (uid=%s): %d", DeviceName(dev.DeviceID), dev.UID, reg.ID)
		}
		return
	}

	reg, err := regFunc(dev)
	if err != nil {
		log.Warnf("failed to register device %s (uid=%s): %s", DeviceName(dev.DeviceID), dev.UID, err)
		return
	}
	if len(reg) == 0 {
		log.Debugf("no registry returned from %s (uid=%s)", DeviceName(dev.DeviceID), dev.UID)
		return
	}
	b.Data.Devices[dev.UID] = dev
	b.Registry[dev.UID] = reg
	for _, reg := range b.Registry[dev.UID] {
		log.Debugf("callback registered for %s (uid=%s): %d", DeviceName(dev.DeviceID), dev.UID, reg.ID)
	}
}
