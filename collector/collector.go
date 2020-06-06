package collector

import (
	"strconv"
	"sync"
	"time"

	"github.com/Tinkerforge/go-api-bindings/ipconnection"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	// bricks:
	"github.com/Tinkerforge/go-api-bindings/master_brick"

	// bricklets:
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_v2_bricklet"
)

// BrickdCollector does all the work
type BrickdCollector struct {
	sync.RWMutex
	Address    string
	Password   string
	Data       *BrickData
	Registry   map[string][]Register
	Connection ipconnection.IPConnection
	Values     chan Value
	Devices    map[uint16]RegisterFunc
}

// RegisterFunc is the funcion of BrickdCollector to register callbacks
type RegisterFunc func(string) []Register

// BrickData are discovered devices and their values
type BrickData struct {
	Address string
	Devices map[string]*Device
	Values  map[string][]Value
}

// Value is returned from the callbacks
type Value struct {
	Index    int                  // index in BrickData.Values, needs to be assigned by the callback
	DeviceID uint16               // https://www.tinkerforge.com/en/doc/Software/Device_Identifier.html
	UID      string               // UID as given from brickd
	Type     prometheus.ValueType // probably just prometheus.GaugeValue
	Help     string               // help for users, i.e. prometheus' "# HELP brickd_humidity_value ..." line, (just the help text)
	Name     string               // value name, such as "usb_voltage" or "humidity"
	Value    float64              // the measurement value
}

// Register is a callback register, the Deregister func will be called as reg.Deregister(reg.ID)
type Register struct {
	Deregister func(uint64)
	ID         uint64
}

// Device is a discovered device
type Device struct {
	UID             string
	ConnectedUID    string
	Position        rune
	HardwareVersion string
	FirmwareVersion string
	DeviceID        uint16
	Available       bool
}

// NewCollector creates a new collector for the given address (and authenticates with the password)
func NewCollector(addr, password string) *BrickdCollector {
	brickd := &BrickdCollector{
		Address:  addr,
		Password: password,
		Data: &BrickData{
			Address: addr,
			Devices: make(map[string]*Device),
			Values:  make(map[string][]Value),
		},
		Registry: make(map[string][]Register),
		Values:   make(chan Value),
	}
	brickd.Devices = map[uint16]RegisterFunc{
		// Bricks
		master_brick.DeviceIdentifier: brickd.RegisterMasterBrick,
		// Bricklets
		barometer_bricklet.DeviceIdentifier:   brickd.RegisterBarometerBricklet,
		humidity_bricklet.DeviceIdentifier:    brickd.RegisterHumidityBricklet,
		humidity_v2_bricklet.DeviceIdentifier: brickd.RegisterHumidityV2Bricklet,
	}

	go brickd.Update()
	return brickd
}

// Update runs in the background and discovers devices and collects the Values
func (b *BrickdCollector) Update() {
	b.Connection = ipconnection.New()
	defer b.Connection.Close()
	b.Connection.SetAutoReconnect(false) // set to true after first successful connection
	b.Connection.RegisterEnumerateCallback(b.OnEnumerate)
	b.Connection.RegisterDisconnectCallback(b.OnDisconnect)
	b.Connection.RegisterConnectCallback(b.OnConnect)

	b.Connection.Connect(b.Address) // Connect to brickd.
	defer b.Connection.Disconnect()

	go func() { // discover eventually new bricks / bricklets on the brickd
		for {
			time.Sleep(time.Minute)
			b.Connection.Enumerate()
		}
	}()

	for v := range b.Values {
		b.Lock()
		log.Debugf("received value from \"%s\" (uid=%s): %s=%f\n", DeviceName(v.DeviceID), v.UID, v.Name, v.Value)
		if _, ok := b.Data.Values[v.UID]; !ok {
			b.Data.Values[v.UID] = make([]Value, 4) // FIXME OutdoorWeather_Bricklet may have more values
		}
		b.Data.Values[v.UID][v.Index] = v
		b.Unlock()
	}
}

// Describe is part of the prometheus.Collector interface
func (b *BrickdCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect is part of the prometheus.Collector interface
func (b *BrickdCollector) Collect(ch chan<- prometheus.Metric) {
	b.RLock()
	defer b.RUnlock()
	for _, vals := range b.Data.Values {
		for _, v := range vals {
			if v.UID == "" {
				continue
			}
			labels := map[string]string{
				"uid":    v.UID,
				"brickd": b.Data.Address,
				"id":     strconv.FormatInt(int64(v.DeviceID), 10),
				"type":   DeviceName(v.DeviceID),
			}

			desc := prometheus.NewDesc(
				"brickd_"+v.Name+"_value", // FIXME do we have anything else than gauge?
				v.Help,
				nil,
				labels,
			)
			ch <- prometheus.MustNewConstMetric(
				desc,
				v.Type,
				v.Value,
			)
		}
	}
}
