package sitlgen

import (
	"types"
)

const (
	/*
		PERM_ANGLE    = 1
		PERM_HORIZON  = 2
		PERM_ALTHOLD  = 3
		PERM_RTH      = 10
		PERM_POSHOLD  = 11
		PERM_MANUAL   = 12
		PERM_MSP_OVER = 13
		PERM_FS       = 27
		PERM_WP       = 28
		PERM_LAUNCH   = 36
		PERM_CRUISE   = 45
	*/
	PERM_ARM              = 0
	PERM_ANGLE            = 1
	PERM_HORIZON          = 2
	PERM_NAV_ALTHOLD      = 3
	PERM_HEADING_HOLD     = 5
	PERM_HEADFREE         = 6
	PERM_HEADADJ          = 7
	PERM_CAMSTAB          = 8
	PERM_NAV_RTH          = 10
	PERM_NAV_POSHOLD      = 11
	PERM_MANUAL           = 12
	PERM_BEEPER           = 13
	PERM_LEDLOW           = 15
	PERM_LIGHTS           = 16
	PERM_OSD_SW           = 19
	PERM_TELEMETRY        = 20
	PERM_AUTO_TUNE        = 21
	PERM_BLACKBOX         = 26
	PERM_FAILSAFE         = 27
	PERM_NAV_WP           = 28
	PERM_AIR_MODE         = 29
	PERM_HOME_RESET       = 30
	PERM_GCS_NAV          = 31
	PERM_FPV_ANGLE_MIX    = 32
	PERM_SURFACE          = 33
	PERM_FLAPERON         = 34
	PERM_TURN_ASSIST      = 35
	PERM_NAV_LAUNCH       = 36
	PERM_SERVO_AUTOTRIM   = 37
	PERM_KILLSWITCH       = 38
	PERM_CAMERA_CONTROL_1 = 39
	PERM_CAMERA_CONTROL_2 = 40
	PERM_CAMERA_CONTROL_3 = 41
	PERM_OSD_ALT_1        = 42
	PERM_OSD_ALT_2        = 43
	PERM_OSD_ALT_3        = 44
	PERM_NAV_CRUISE       = 45
	PERM_MC_BRAKING       = 46
	PERM_USER1            = 47
	PERM_USER2            = 48
	PERM_LOITER_CHANGE    = 49
	PERM_MSP_RC_OVERRIDE  = 50
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
	{types.FM_LAUNCH, []uint16{PERM_NAV_LAUNCH}, "Launch"},
	{types.FM_RTH, []uint16{PERM_NAV_RTH}, "RTH"},
	{types.FM_WP, []uint16{PERM_NAV_WP}, "WP"},
	{types.FM_CRUISE3D, []uint16{PERM_NAV_CRUISE}, "Cruise3D"},
	{types.FM_CRUISE2D, []uint16{PERM_NAV_CRUISE}, "Cruise2D"},
	{types.FM_PH, []uint16{PERM_NAV_POSHOLD}, "PosHold"},
	{types.FM_AH, []uint16{PERM_NAV_ALTHOLD}, "AltHold"},
	{types.FM_EMERG, []uint16{0xfffe}, "Emergency"},
	{types.FM_FS, []uint16{PERM_FAILSAFE}, "Failsafe"},
	{types.FM_ARM, []uint16{PERM_ARM}, "Arm"},
	{types.FM_MSP_OVER, []uint16{PERM_MSP_RC_OVERRIDE}, "Override"},
	{types.FM_BEEPER, []uint16{PERM_BEEPER}, "Beeper"},
	{types.FM_GCS_NAV, []uint16{PERM_GCS_NAV}, "GCS Nav"},
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
