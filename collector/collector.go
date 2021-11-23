package collector

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/Tinkerforge/go-api-bindings/air_quality_bricklet"
	"github.com/Tinkerforge/go-api-bindings/ipconnection"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	// bricks:
	"github.com/Tinkerforge/go-api-bindings/hat_zero_brick"
	"github.com/Tinkerforge/go-api-bindings/master_brick"

	// bricklets:
	"github.com/Tinkerforge/go-api-bindings/barometer_bricklet"
	"github.com/Tinkerforge/go-api-bindings/barometer_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_bricklet"
	"github.com/Tinkerforge/go-api-bindings/humidity_v2_bricklet"
	"github.com/Tinkerforge/go-api-bindings/outdoor_weather_bricklet"
)

const (
	EthernetCallbackID uint64 = math.MaxUint64
)

// BrickdCollector does all the work
type BrickdCollector struct {
	sync.RWMutex
	Address        string
	Password       string
	Data           *BrickData
	Registry       map[string][]Register
	Connection     ipconnection.IPConnection
	Values         chan Value
	Devices        map[uint16]RegisterFunc
	CallbackPeriod uint32
	IgnoredUIDs    []string
	Labels         map[string]string
	SensorLabels   map[string]map[string]map[string]string
	EthernetState  chan interface{}
}

// RegisterFunc is the funcion of BrickdCollector to register callbacks
type RegisterFunc func(string) ([]Register, error)

// BrickData are discovered devices and their values
type BrickData struct {
	Address string
	Devices map[string]*Device
	Values  map[string]map[int]Value
}

// Value is returned from the callbacks
type Value struct {
	Index    int                  // index in BrickData.Values, needs to be assigned by the callback
	DeviceID uint16               // https://www.tinkerforge.com/en/doc/Software/Device_Identifier.html
	UID      string               // UID as given from brickd
	SensorID int                  // sensor id in outdoor_weather_bricklet
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
func NewCollector(addr, password string, cbPeriod time.Duration, ignoredUIDs []string, labels map[string]string, sensorLabels map[string]map[string]map[string]string) *BrickdCollector {
	brickd := &BrickdCollector{
		Address:  addr,
		Password: password,
		Data: &BrickData{
			Address: addr,
			Devices: make(map[string]*Device),
			Values:  make(map[string]map[int]Value),
		},
		Registry:       make(map[string][]Register),
		Values:         make(chan Value),
		CallbackPeriod: uint32(cbPeriod / time.Millisecond),
		IgnoredUIDs:    ignoredUIDs,
		Labels:         labels,
		SensorLabels:   sensorLabels,
	}
	brickd.Devices = map[uint16]RegisterFunc{
		// Bricks
		master_brick.DeviceIdentifier:   brickd.RegisterMasterBrick,
		hat_zero_brick.DeviceIdentifier: brickd.RegisterZeroHatBrick,

		// Bricklets
		air_quality_bricklet.DeviceIdentifier:  brickd.RegisterAirQualityBricklet,
		barometer_bricklet.DeviceIdentifier:    brickd.RegisterBarometerBricklet,
		barometer_v2_bricklet.DeviceIdentifier: brickd.RegisterBarometerBricklet,
		humidity_bricklet.DeviceIdentifier:     brickd.RegisterHumidityBricklet,
		humidity_v2_bricklet.DeviceIdentifier:  brickd.RegisterHumidityV2Bricklet,

		outdoor_weather_bricklet.DeviceIdentifier: brickd.RegisterOutdoorWeatherBricklet,
	}

	go brickd.Update()
	return brickd
}

func (b *BrickdCollector) ignored(uid string) bool {
	for _, u := range b.IgnoredUIDs {
		if uid == u {
			return true
		}
	}
	return false
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
		if b.ignored(v.UID) {
			continue
		}
		b.Lock()
		log.Debugf("received value from \"%s\" (uid=%s, sensor=%d): %s=%f\n", DeviceName(v.DeviceID), v.UID, v.SensorID, v.Name, v.Value)
		if _, ok := b.Data.Values[v.UID]; !ok {
			b.Data.Values[v.UID] = make(map[int]Value)
		}
		b.Data.Values[v.UID][v.Index] = v
		b.Unlock()
		// log.Debugf("DATA=%#v", b.Data.Values)
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
			if v.UID == "" || b.ignored(v.UID) {
				continue
			}
			labels := map[string]string{
				"uid":       v.UID,
				"brickd":    b.Data.Address,
				"id":        strconv.FormatInt(int64(v.DeviceID), 10),
				"type":      DeviceName(v.DeviceID),
				"sub_id":    strconv.Itoa(v.SensorID), // deprecated
				"sensor_id": strconv.Itoa(v.SensorID),
			}
			for k, v := range b.Labels {
				if _, exists := labels[k]; exists {
					continue
				}
				labels[k] = v
			}
			if sl, ok := b.SensorLabels[v.UID]; ok {
				if l, ok := sl[strconv.Itoa(v.SensorID)]; ok {
					for k, v := range l {
						if _, exists := labels[k]; exists {
							continue
						}
						labels[k] = v
					}
				}
			}

			var promType string
			switch v.Type {
			case prometheus.CounterValue:
				promType = "total"
			case prometheus.GaugeValue:
				promType = "value"
			}
			desc := prometheus.NewDesc(
				"brickd_"+v.Name+"_"+promType,
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
