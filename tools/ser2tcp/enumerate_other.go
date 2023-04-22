//go:build !linux
// +build !linux

package main

import (
	"go.bug.st/serial/enumerator"
)

func get_device_by_description(desc string) string {
	devname := ""
	ports, err := enumerator.GetDetailedPortsList()
	if err == nil {
		for _, port := range ports {
			devname = port.Name
			if port.Product == desc {
				return devname
			}
		}
	}
	if devname != "" {
		return devname
	} else {
		return desc
	}
}
