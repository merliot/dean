//go:build m5stack_core2

package main

import (
	"machine"

	"github.com/merliot/dean/drivers/ft6336"
	"github.com/merliot/dean/drivers/i2csoft"
	"github.com/merliot/dean/drivers/touch"
)

// InitDisplay initializes the display of each board.
func initDevices() (touch.Pointer, error) {
	i2c := i2csoft.New(machine.SCL0_PIN, machine.SDA0_PIN)
	i2c.Configure(i2csoft.I2CConfig{Frequency: 100e3})

	resistiveTouch := ft6336.New(i2c, machine.Pin(39))
	resistiveTouch.Configure(ft6336.Config{})
	resistiveTouch.SetPeriodActive(0x00)

	return resistiveTouch, nil
}
