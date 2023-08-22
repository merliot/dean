# TinyGo Drivers

[![PkgGoDev](https://pkg.go.dev/badge/github.com/merliot/dean/drivers)](https://pkg.go.dev/github.com/merliot/dean/drivers) [![Build](https://github.com/tinygo-org/drivers/actions/workflows/build.yml/badge.svg?branch=dev)](https://github.com/tinygo-org/drivers/actions/workflows/build.yml)


This package provides a collection of hardware drivers for devices such as sensors and displays that can be used together with [TinyGo](https://tinygo.org).

## Installing

```shell
go get github.com/merliot/dean/drivers
```

## How to use

Here is an example in TinyGo that uses the BMP180 digital barometer:

```go
package main

import (
    "time"

    "machine"

    "github.com/merliot/dean/drivers/bmp180"
)

func main() {
    machine.I2C0.Configure(machine.I2CConfig{})
    sensor := bmp180.New(machine.I2C0)
    sensor.Configure()

    connected := sensor.Connected()
    if !connected {
        println("BMP180 not detected")
        return
    }
    println("BMP180 detected")

    for {
        temp, _ := sensor.ReadTemperature()
        println("Temperature:", float32(temp)/1000, "°C")

        pressure, _ := sensor.ReadPressure()
        println("Pressure", float32(pressure)/100000, "hPa")

        time.Sleep(2 * time.Second)
    }
}
```

## Supported devices

There are currently 96 devices supported. For the complete list, please see:
https://tinygo.org/docs/reference/devices/

## Contributing

Your contributions are welcome!

Please take a look at our [CONTRIBUTING.md](./CONTRIBUTING.md) document for details.

## License

This project is licensed under the BSD 3-clause license, just like the [Go project](https://golang.org/LICENSE) itself.
