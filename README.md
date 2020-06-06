# Prometheus Brickd Exporter

The brickd exporter is a Prometheus exporter which connects to a [Tinkerforge](https://www.tinkerforge.com/)
[brickd](https://www.tinkerforge.com/en/doc/Software/Brickd.html) and exports the values from the
connected bricks and bricklets.

Data from the brickd is collected in the background. Current *CallbackPeriod are set to 10,000 ms (likely
to be configurable in the future).

## Usage

### Pre-requisite

go version 1.11 or later (go modules are used in this repo).

### Building the brickd exporter

Clone the repository via `go get github.com/vetinari/brickd_exporter` or git and cd into the directory.
Then build the brickd_exporter binary with the following commands.

    $ go build

### Configuration

When no `--config.file /path/to/brickd.yml` is given, the default config is:

```yaml
collector:
    log_level: info
listen:
    address: :9639
    metrics_path: /metrics
brickd:
    address: localhost:4223
```

Any of these values can be set. Use the default `brickd.address` when the bricks are connected
on local USB port.

`collector.log_level` can be set to `debug` to see the devices discovered and their values received
from the callbacks.

### Running

Start with `--config.file /path/to/brickd.yml` to pass a config file. 

## Suported bricks and bricklets

Bricks:

* Master Brick

Bricklets:

* Humidity Bricklet

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
