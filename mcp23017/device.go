// Package mcp23017 implements a driver for the MCP23017
// I2C port expander chip. See https://www.microchip.com/wwwproducts/en/MCP23017
// for details of the interface.
package mcp23017

import (
	"errors"
	"machine"
)

const (
	// hwAddressFixed holds the bits of the hardware address
	// that are fixed by the chip. Bits 0-3 (those in hwAddressMask)
	// are user-defined by the A0-A2 pins on the chip.
	hwAddress = uint8(0b010_0000)
	// hwAddressMask holds the bits that are significant in hwAddress.
	hwAddressMask = uint8(0b111_1000)
)

type register uint8

const (
	// The following registers all refer to port A (except
	// rIOCON with is port-agnostic).
	// ORing them with portB makes them refer to port B.
	rIODIR   = register(0x00) // I/O direction. 0=output; 1=input.
	rIOPOL   = register(0x02) // Invert input values. 0=normal; 1=inverted.
	rGPINTEN = register(0x04)
	rDEFVAL  = register(0x06)
	rINTCON  = register(0x08)
	rIOCON   = register(0x0A)
	rGPPU    = register(0x0C) // Pull up; 1=pull-up.
	rINTF    = register(0x0E)
	rINTCAP  = register(0x10)
	rGPIO    = register(0x12) // GPIO pin values.
	rOLAT    = register(0x14)

	portB = register(0x1)
)

// PinCount is the number of GPIO pins available on the chip.
const PinCount = 16

// PinMode represents a possible I/O mode for a pin.
// The zero value represents the default value
// after the chip is reset (input).
type PinMode uint8

const (
	// Input configures a pin as an input.
	Input = PinMode(0)
	// Output configures a pin as an output.
	Output = PinMode(1)

	// Direction is the bit mask of the pin mode representing
	// the I/O direction.
	Direction = PinMode(1)

	// Pullup can be bitwise-or'd with Input
	// to cause the pull-up resistor on the pin to
	// be enabled.
	Pullup = PinMode(2)

	// Invert can be bitwise-or'd with Input to
	// cause the pin value to reflect the inverted
	// value on the pin.
	Invert = PinMode(4)
)

var ErrInvalidHWAddress = errors.New("invalid hardware address")

// New returns a new MCP23017 device at the given I2C address.
// It returns ErrInvalidHWAddress if the address isn't possible for the device.
//
// By default all pins are configured as inputs.
func NewI2C(bus machine.I2C, address uint8) (*Device, error) {
	if address&hwAddressMask != hwAddress {
		return nil, ErrInvalidHWAddress
	}
	d := &Device{
		bus:  bus,
		addr: address,
	}
	_, err := d.readRegister(rGPIO)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Device represents an MCP23017 device.
type Device struct {
	bus  machine.I2C
	addr uint8
}

// GetPins reads all 16 pins from ports A and B.
func (d *Device) GetPins() (Pins, error) {
	return d.readRegisterAB(rGPIO)
}

// SetPins writes all the pins at once. The bits
// are laid out as for GetPins.
func (d *Device) SetPins(pins Pins) error {
	return d.writeRegisterAB(rGPIO, pins)
}

// Pin returns a Pin representing the given pin number (from 0 to 15).
// Pin numbers from 0 to 7 represent port A pins 0 to 7.
// Pin numbers from 8 to 15 represent port B pins 0 to 7.
func (d *Device) Pin(pin int) Pin {
	if pin < 0 || pin >= PinCount {
		panic("pin out of range")
	}
	var port register
	if pin > 7 {
		port = portB
	}
	return Pin{
		dev:  d,
		port: port,
		mask: uint8(1 << (pin & 0x7)),
		pin:  uint8(pin),
	}
}

// SetAllModes sets the mode of all the pins in a single operation.
func (d *Device) SetModes(modes *[PinCount]PinMode) error {
	var dir, pullup, invert Pins
	for i, mode := range modes {
		if mode&Direction == Input {
			dir.High(i)
		}
		if mode&Pullup != 0 {
			pullup.High(i)
		}
		if mode&Invert != 0 {
			invert.High(i)
		}
	}
	if err := d.writeRegisterAB(rIODIR, dir); err != nil {
		return err
	}
	if err := d.writeRegisterAB(rGPPU, pullup); err != nil {
		return err
	}
	if err := d.writeRegisterAB(rIOPOL, invert); err != nil {
		return err
	}
	return nil
}

// GetModes reads the modes of all the pins into modes.
func (d *Device) GetModes(modes *[PinCount]PinMode) error {
	dir, err := d.readRegisterAB(rIODIR)
	if err != nil {
		return err
	}
	pullup, err := d.readRegisterAB(rGPPU)
	if err != nil {
		return err
	}
	invert, err := d.readRegisterAB(rIOPOL)
	if err != nil {
		return err
	}
	for i := range modes {
		mode := Output
		if dir.Get(i) {
			mode = Input
		}
		if pullup.Get(i) {
			mode |= Pullup
		}
		if invert.Get(i) {
			mode |= Invert
		}
		modes[i] = mode
	}
	return nil
}

func (d *Device) readRegister(r register) (uint8, error) {
	var buf [1]byte
	if err := d.bus.ReadRegister(d.addr, uint8(r), buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (d *Device) writeRegister(r register, val uint8) error {
	buf := [1]byte{val}
	return d.bus.WriteRegister(d.addr, uint8(r), buf[:])
}

func (d *Device) writeRegisterAB(r register, val Pins) error {
	// We rely on the auto-incrementing sequential write
	// and the fact that registers alternate between A and B
	// to write both ports in a single operation.
	buf := [2]byte{uint8(val), uint8(val >> 8)}
	return d.bus.WriteRegister(d.addr, uint8(r&^portB), buf[:])
}

func (d *Device) readRegisterAB(r register) (Pins, error) {
	// We rely on the auto-incrementing sequential write
	// and the fact that registers alternate between A and B
	// to read both ports in a single operation.
	var buf [2]byte
	if err := d.bus.ReadRegister(d.addr, uint8(r), buf[:]); err != nil {
		return Pins(0), err
	}
	return Pins(buf[0]) | (Pins(buf[1]) << 8), nil
}

// Pin represents a single GPIO pin on the device.
type Pin struct {
	// mask holds the mask of the pin within its specific register.
	mask uint8
	// port holds the register bank to use (portB or 0)
	port register
	// pin holds the actual pin number.
	pin uint8
	dev *Device
}

// Set sets the pin to the given value.
func (p Pin) Set(value bool) error {
	r := rGPIO | p.port
	v, err := p.dev.readRegister(r)
	if err != nil {
		return err
	}
	if value {
		v |= p.mask
	} else {
		v &^= p.mask
	}
	return p.dev.writeRegister(r, v)
}

// High is short for p.Set(true).
func (p Pin) High() error {
	return p.Set(true)
}

// High is short for p.Set(false).
func (p Pin) Low() error {
	return p.Set(true)
}

// Get returns the current value of the given pin.
func (p Pin) Get() (bool, error) {
	v, err := p.dev.readRegister(rGPIO | p.port)
	if err != nil {
		return false, err
	}
	return v&p.mask != 0, nil
}

// SetMode configures the pin to the given mode.
func (p Pin) SetMode(mode PinMode) error {
	// We could use a more efficient single-register
	// read/write pattern but setting pin modes isn't an
	// operation that's likely to need to be efficient, so
	// use less code and use Get/SetModes directly.
	var modes [PinCount]PinMode
	if err := p.dev.GetModes(&modes); err != nil {
		return err
	}
	modes[p.pin] = mode
	return p.dev.SetModes(&modes)
}

// GetMode returns the mode of the pin.
func (p Pin) GetMode() (PinMode, error) {
	var modes [PinCount]PinMode
	if err := p.dev.GetModes(&modes); err != nil {
		return 0, err
	}
	return modes[p.pin], nil
}

// Pins represents a bitmask of pin values.
// Port A values are in bits 0-8 (numbered from least significant bit)
// Port B values are in bits 9-15.
type Pins uint16

// Set sets the value for the given pin.
func (p *Pins) Set(pin int, value bool) {
	if value {
		p.High(pin)
	} else {
		p.Low(pin)
	}
}

// Get returns the value for the given pin.
func (p Pins) Get(pin int) bool {
	return (p & pinMask(pin)) != 0
}

// High is short for p.Set(pin, true).
func (p *Pins) High(pin int) {
	*p |= pinMask(pin)
}

// Low is short for p.Set(pin, false).
func (p *Pins) Low(pin int) {
	*p &^= pinMask(pin)
}

func pinMask(pin int) Pins {
	return 1 << pin
}
