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
		println("cannot configure i2c: ", err.Error())
		return
	}

	addrs := []uint8{0x20, 0x21, 0x22}
	mcp := make([]*mcp23017.Device, len(addrs))
	for i, addr := range addrs {
		d, err := mcp23017.NewI2C(machine.I2C0, addr)
		if err != nil {
			println("cannot make new i2c for ", addr, err.Error())
			return
		}
		mcp[i] = d
	}
	setAllMode(mcp[0], mcp23017.Output)
	setAllMode(mcp[1], mcp23017.Output)
	setAllMode(mcp[2], mcp23017.Input|mcp23017.Pullup|mcp23017.Invert)
	buttonc := make(chan uint8)
	go blinkenlights(mcp[0:2], buttonc)
	go func() {
		if err := buttons(mcp[2], buttonc); err != nil {
			println("cannot read buttons: ", err)
		}
	}()
	select {}
}

func blinkenlights(mcp []*mcp23017.Device, buttonc <-chan uint8) {
	x := uint32(0)
	//0b1010_1111_1100_0111_1000_0001_0001_1000)
	tickc := make(chan struct{})
	go ticker(200*time.Millisecond, tickc)
	for {
		mcp[0].SetPins(mcp23017.Pins(x))
		mcp[1].SetPins(mcp23017.Pins(x >> 16))
		select {
		case b := <-buttonc:
			x ^= uint32(b)
			continue
		case <-tickc:
		}
		x = rotateLeft24(x)
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

func setAllMode(d *mcp23017.Device, mode mcp23017.PinMode) error {
	var modes [mcp23017.PinCount]mcp23017.PinMode
	for i := range modes {
		modes[i] = mode
	}
	if err := d.SetModes(&modes); err != nil {
		return err
	}
	return nil
}
