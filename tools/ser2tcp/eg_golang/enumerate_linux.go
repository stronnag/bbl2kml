//go:build linux && amd64

package main

import (
	"github.com/jochenvg/go-udev"
)

func get_device_by_description(desc string) string {
	devname := ""
	u := udev.Udev{}
	e := u.NewEnumerate()
	e.AddMatchSubsystem("tty")
	e.AddMatchProperty("ID_BUS", "usb")
	devices, _ := e.Devices()
	for _, d := range devices {
		dp := d.Properties()
		devname = dp["DEVNAME"]
		if dp["ID_USB_MODEL"] == desc {
			return devname
		}
	}
	if devname != "" {
		return devname
	}
	return desc
}
