// +build tinygo

package main

import (
	"machine"

	"github.com/rogpeppe/doorbell/mcp23017"
)

func getDevices(addrs ...uint8) (mcp23017.Devices, error) {
	if err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	}); err != nil {
		fatal("cannot configure i2c: ", err.Error())
	}
	return mcp23017.NewI2CDevices(machine.I2C0, addrs...)
}
