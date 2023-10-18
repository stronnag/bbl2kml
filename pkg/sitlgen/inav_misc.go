package sitlgen

import (
	"types"
)

const (
	PERM_ARM     = 0
	PERM_MANUAL  = 12
	PERM_HORIZON = 2
	PERM_ANGLE   = 1
	PERM_LAUNCH  = 36
	PERM_RTH     = 10
	PERM_WP      = 28
	PERM_CRUISE  = 45
	PERM_ALTHOLD = 3
	PERM_POSHOLD = 11
	PERM_FS      = 27
)

type FModeMap struct {
	fmode  uint16
	imodes []uint16
	name   string
}

var fmodes []FModeMap = []FModeMap{
	{types.FM_ACRO, []uint16{0xffff}, "Acro"},
	{types.FM_MANUAL, []uint16{PERM_MANUAL}, "Manual"},
	{types.FM_HORIZON, []uint16{PERM_HORIZON}, "Horizon"},
	{types.FM_ANGLE, []uint16{PERM_ANGLE}, "Angle"},
	{types.FM_LAUNCH, []uint16{PERM_LAUNCH}, "Launch"},
	{types.FM_RTH, []uint16{PERM_RTH}, "RTH"},
	{types.FM_WP, []uint16{PERM_WP}, "WP"},
	{types.FM_CRUISE3D, []uint16{PERM_CRUISE}, "Cruise3D"},
	{types.FM_CRUISE2D, []uint16{PERM_CRUISE}, "Cruise2D"},
	{types.FM_PH, []uint16{PERM_POSHOLD}, "PosHold"},
	{types.FM_AH, []uint16{PERM_ALTHOLD}, "AltHold"},
	{types.FM_EMERG, []uint16{0xfffe}, "Emergency"},
	{types.FM_FS, []uint16{PERM_FS}, "Failsafe"},
	{types.FM_ARM, []uint16{PERM_ARM}, "Arm"},
}

func fm_to_mode(fm uint16) ([]uint16, string) {
	for _, m := range fmodes {
		if m.fmode == fm {
			return m.imodes, m.name
		}
	}
	return []uint16{}, ""
}

func mode_to_fm(imode uint16) (uint16, string) {
	for _, m := range fmodes {
		for _, i := range m.imodes {
			if imode == i {
				return m.fmode, m.name
			}
		}
	}
	return 0, ""
}
