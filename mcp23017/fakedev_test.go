package mcp23017

import (
	qt "github.com/frankban/quicktest"
)

type fakeBus struct {
	c    *qt.C
	devs []*fakeDev
}

func newBus(c *qt.C) *fakeBus {
	return &fakeBus{
		c: c,
	}
}

type fakeDev struct {
	c         *qt.C
	addr      uint8
	Registers [registerCount]uint8
	Err       error
}

func (bus *fakeBus) addDevice(addr uint8) *fakeDev {
	dev := &fakeDev{
		c:    bus.c,
		addr: addr,
		Registers: [registerCount]uint8{
			// IODIRA and IODIRB are all ones by default.
			rIODIR:         0xff,
			rIODIR | portB: 0xff,
		},
	}
	bus.devs = append(bus.devs, dev)
	return dev
}

func (bus *fakeBus) ReadRegister(addr uint8, r uint8, buf []byte) error {
	return bus.findDev(addr).ReadRegister(r, buf)
}

func (bus *fakeBus) WriteRegister(addr uint8, r uint8, buf []byte) error {
	return bus.findDev(addr).WriteRegister(r, buf)
}

func (d *fakeDev) ReadRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(buf, d.Registers[r:])
	return nil
}

func (d *fakeDev) WriteRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(d.Registers[r:], buf)
	return nil
}

func (d *fakeDev) assertRegisterRange(r uint8, buf []byte) {
	if int(r) >= len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)+len(buf) > len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}

func (bus *fakeBus) findDev(addr uint8) *fakeDev {
	for _, dev := range bus.devs {
		if dev.addr == addr {
			return dev
		}
	}
	bus.c.Fatalf("invalid device addr %#x passed to i2c bus", addr)
	panic("unreachable")
}
