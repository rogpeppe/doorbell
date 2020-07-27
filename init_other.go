// +build !tinygo

package main

import "github.com/rogpeppe/doorbell/mcp23017"

func getDevices(addrs ...uint8) (mcp23017.Devices, error) {
	panic("this only runs with tinygo")
}
