package main

import (
	"machine"
	"time"

	"github.com/rogpeppe/doorbell/mcp23017"
)

func main() {
	time.Sleep(3 * time.Second)
	println("testing....")
	if err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	}); err != nil {
		fatal("cannot configure i2c: ", err.Error())
	}

	devs, err := mcp23017.NewI2CDevices(machine.I2C0, 0x20, 0x21, 0x22)
	if err != nil {
		fatal("cannot make devices", err.Error())
	}
	outputs := devs[0:2]
	outputs.SetModes([]mcp23017.PinMode{mcp23017.Output})
	inputs := devs[2]
	inputs.SetModes([]mcp23017.PinMode{mcp23017.Input | mcp23017.Pullup | mcp23017.Invert})
	buttonc := make(chan uint8)
	go blinkenlights(outputs, buttonc)
	go func() {
		if err := buttons(inputs, buttonc); err != nil {
			println("cannot read buttons: ", err)
		}
	}()
	select {}
}

func blinkenlights(devs mcp23017.Devices, buttonc <-chan uint8) {
	x := uint32(0)
	tickc := make(chan struct{})
	go ticker(500*time.Millisecond, tickc)
	pins := make(mcp23017.PinSlice, 2)
	mask := uint32(1<<24 - 1)
	n := uint(0)
	for {
		shift := n % 24
		shifted := ((x << shift) | (x >> (24 - shift))) & mask

		pins[0] = mcp23017.Pins(shifted)
		pins[1] = mcp23017.Pins(shifted >> 16)
		devs.SetPins(pins, mcp23017.All)
		select {
		case b := <-buttonc:
			x = uint32(b)
			continue
		case <-tickc:
		}
		n++
	}
}

func rotateLeft24(x uint32) uint32 {
	x &= 0xffffff
	return (x<<1 | x>>23) & 0xffffff
}

func ticker(interval time.Duration, c chan<- struct{}) {
	for {
		c <- struct{}{}
		time.Sleep(interval)
	}
}

func buttons(d *mcp23017.Device, buttonc chan<- uint8) error {
	b, err := d.GetPins()
	if err != nil {
		return err
	}
	for {
		time.Sleep(time.Millisecond)
		b1, err := d.GetPins()
		if err != nil {
			return err
		}
		if b1 != b {
			b = b1
			buttonc <- uint8(b)
		}
	}
}

func fatal(args ...interface{}) {
	print("fatal: ")
	for _, a := range args {
		print(a)
	}
	println()
	select {}
}
