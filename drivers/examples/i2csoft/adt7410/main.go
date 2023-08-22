package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/dean/drivers/adt7410"
	"github.com/merliot/dean/drivers/i2csoft"
)

func main() {
	i2c := i2csoft.New(machine.SCL_PIN, machine.SDA_PIN)
	i2c.Configure(i2csoft.I2CConfig{
		Frequency: 400e3,
	})

	sensor := adt7410.New(i2c)
	sensor.Configure()

	for {
		temp := sensor.ReadTempF()
		fmt.Printf("temperature: %f\r\n", temp)
		time.Sleep(time.Second)
	}

}
