# Prometheus Brickd Exporter

The brickd exporter is a Prometheus exporter which connects to a [Tinkerforge](https://www.tinkerforge.com/)
[brickd](https://www.tinkerforge.com/en/doc/Software/Brickd.html) and exports the values from the
connected bricks and bricklets. It can additionally send the data to an MQTT broker, see [MQTT](#mqtt) section
on how to configure the broker and topics.

Data from the brickd is collected in the background. Currently callbackPeriod is set to 10,000 ms, which can
be changed in the config.

Note: the `sub_id` label has been deprecated, it is replaced by the `sensor_id` label in the metrics and
will be removed in the future.

## Usage

### Pre-requisite

go version 1.21 or later.

### Building the brickd exporter

Clone the repository via `go get github.com/vetinari/brickd_exporter` or git and cd into the directory.
Then build the brickd\_exporter binary with the following commands.

    $ make build

or to build for a raspberry pi:

    $ make build.raspi

The resulting binary will be in the `./build/` directory

### Configuration

When no `--config.file /path/to/brickd.yml` is given, the default config is:

```yaml
collector:
    log_level: info
    callback_period: 10s
    ignored_uids: []
    led_status: "on"
    expire_period: 0s
listen:
    address: :9639
    metrics_path: /metrics
brickd:
    address: localhost:4223
mqtt:
    enabled: false
    topic: "brickd/"
    homeassistant:
       enabled: false
       discovery_base: homeassistant/
       interval: 5m
```

Any of these values can be set. Use the default `brickd.address` when the bricks are connected
on local USB port.

`collector.log_level` can be set to `debug` to see the devices discovered and their values received
from the callbacks.

`collector.labels` is a key -> value map of strings which will be applied to all metrics.

`collector.sensor_labels` is a mapping of the UID of the brick(let), to sensor id (as string, usually
`"0"` for all except with the "Outdoor Weather Bricklet" to a key -> value map of strings, see [brickd.yml](brickd.yml)
for examples. Those will only applied to the defined sensors.

`collector.expire_period` sets a duration after which old values are not exported anymore, i.e. if the latest value of a 
brick / bricklet has been received from brickd more than this period ago it will not be shown anymore. `0s` (or any other
`time.Duration` of `0` disables this feature (the default). Do not set this too low or you might not export anything :) 
Depending on your use case 2 or more times the `collector.callback_period` should be OK.

Setting `collector.led_status` attempts to set the LED of the bricklets where available, default is `"on"`, other possible
values are `"off", `"heartbeat"` and `"status"`.

### MQTT

The MQTT broker is configured in the *mqtt* section. An example config looks like:
```yaml
mqtt:
  enabled: true
  broker:
    host: 192.168.5.33
    port: 1883
    username: brickd
    password: brickd_pass
    client_id: brickd_exporter
  topic: brickd/
```

**Note**: if you're running multiple brickd exporter each one must get a unique `client_id`.

The `mqtt.topic` sets the base topic where each metric is reported to. The target topic (key 
`mqtt_topic`) for the metrics are per device uid + sensor id configured in the `collector.sensor_labels`.
Check the supplied [example config](brickd.yml) how this is done. Note: the `mqtt_topic` will
not be in the labels (not in prometheus and not in the MQTT payload).


The "Master Brick", "HAT Brick" and "HAT Zero Brick" values are reported in the topics `master_brick`, 
`hat_brick` and `hat_zero_brick` topics respectively (prefixed by `mqtt.topic` of course).

### Home Assistant

With MQTT enabled, you can also enable the auto discovery options for
[Home Assistant](https://www.home-assistant.io/). This is done by extending the `mqtt` settings like

```yaml
mqtt:
  enabled: true
  broker:
    host: 192.168.5.33
    port: 1883
    username: brickd
    password: brickd_pass
    client_id: brickd_exporter
  topic: brickd/
  homeassistant:
    enabled: true
    discovery_base: homeassistant/
    interval: 5m
```

After starting, the new devices - one per bricklet - and their entities should show up in your HA setup.

### Running

Start with `--config.file /path/to/brickd.yml` to pass a config file. 

## Suported bricks and bricklets

Bricks:

* [Master Brick](https://www.tinkerforge.com/en/doc/Hardware/Bricks/Master_Brick.html)
* [Zero Hat Brick](https://www.tinkerforge.com/de/doc/Hardware/Bricks/HAT_Zero_Brick.html)
* [Hat Brick](https://www.tinkerforge.com/en/doc/Hardware/Bricks/HAT_Brick.html)

Bricklets:

* [Ambient Light Bricklet 3.0](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Ambient_Light_V3.html)
* [Analog In V3 Bricklet](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Analog_In_V3.html)
* [AirQuality Bricklet](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Air_Quality.html)
* [Barometer Bricklet](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Barometer.html)
* [Barometer Bricklet v2.0](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Barometer_V2.html)
* [CO2 Bricklet 2.0](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/CO2_V2.html)
* [Humidity Bricklet](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Humidity.html)
* [Humidity Bricklet v2.0](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Humidity_V2.html)
* [Outdoor Weather Bricklet](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/Outdoor_Weather.html)
* [UV Light Bricklet v2.0](https://www.tinkerforge.com/en/doc/Hardware/Bricklets/UV_Light_V2.html)

Adding more is easy, see [Contributing](#contributing)

## Contributing

If you would like to contribute code or documentation, follow these steps:

* Clone a local copy.
* Make your changes on a uniquely named branch.
* Comment those changes.
* Test those changes 
* Make sure the code is go formatted (hint: `gofmt -w $file`)
* Push your branch to a fork and create a Pull Request.

### Adding new bricks and bricklets

* add register function, check the existing ones in collector/bricks.go / collector/bricklets.go
  as examples.
* add the functions to the `NewCollector`, example of existing ones:
```go
      brickd.Devices = map[uint16]RegisterFunc{
        master_brick.DeviceIdentifier:      brickd.RegisterMasterBrick,
        humidity_bricklet.DeviceIdentifier: brickd.RegisterHumidityBricklet,
      }
```
* don't forget to update the imports
* test new devices and create a pull request (see above).
