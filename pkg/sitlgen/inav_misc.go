package sitlgen

import (
	"types"
)

const (
	PERM_ARM                = 0
	PERM_ANGLE              = 1
	PERM_HORIZON            = 2
	PERM_NAV_ALTHOLD        = 3
	PERM_HEADING_HOLD       = 5
	PERM_HEADFREE           = 6
	PERM_HEADADJ            = 7
	PERM_CAMSTAB            = 8
	PERM_NAV_RTH            = 10
	PERM_NAV_POSHOLD        = 11
	PERM_MANUAL             = 12
	PERM_BEEPER             = 13
	PERM_LEDS_OFF           = 15
	PERM_LIGHTS             = 16
	PERM_OSD_OFF            = 19
	PERM_TELEMETRY          = 20
	PERM_AUTO_TUNE          = 21
	PERM_BLACKBOX           = 26
	PERM_FAILSAFE           = 27
	PERM_NAV_WP             = 28
	PERM_AIR_MODE           = 29
	PERM_HOME_RESET         = 30
	PERM_GCS_NAV            = 31
	PERM_FPV_ANGLE_MIX      = 32
	PERM_SURFACE            = 33
	PERM_FLAPERON           = 34
	PERM_TURN_ASSIST        = 35
	PERM_NAV_LAUNCH         = 36
	PERM_SERVO_AUTOTRIM     = 37
	PERM_CAMERA_CONTROL_1   = 39
	PERM_CAMERA_CONTROL_2   = 40
	PERM_CAMERA_CONTROL_3   = 41
	PERM_OSD_ALT_1          = 42
	PERM_OSD_ALT_2          = 43
	PERM_OSD_ALT_3          = 44
	PERM_NAV_COURSE_HOLD    = 45
	PERM_MC_BRAKING         = 46
	PERM_LOITER_CHANGE      = 49
	PERM_MSP_RC_OVERRIDE    = 50
	PERM_PREARM             = 51
	PERM_TURTLE             = 52
	PERM_NAV_CRUISE         = 53
	PERM_AUTO_LEVEL_TRIM    = 54
	PERM_WP_PLANNER         = 55
	PERM_SOARING            = 56
	PERM_MISSION_CHANGE     = 59
	PERM_BEEPER_MUTE        = 60
	PERM_MULTI_FUNCTION     = 61
	PERM_MIXER_PROFILE_2    = 62
	PERM_MIXER_TRANSITION   = 63
	PERM_ANGLE_HOLD         = 64
	PERM_GIMBAL_LEVEL_TILT  = 65
	PERM_GIMBAL_LEVEL_ROLL  = 66
	PERM_GIMBAL_CENTER      = 67
	PERM_GIMBAL_HEADTRACKER = 68
)

var pnames = []struct {
	permid uint8
	name   string
}{
	{permid: PERM_ARM, name: "ARM"},
	{permid: PERM_ANGLE, name: "ANGLE"},
	{permid: PERM_HORIZON, name: "HORIZON"},
	{permid: PERM_NAV_ALTHOLD, name: "NAV ALTHOLD"},
	{permid: PERM_HEADING_HOLD, name: "HEADING HOLD"},
	{permid: PERM_HEADFREE, name: "HEADFREE"},
	{permid: PERM_HEADADJ, name: "HEADADJ"},
	{permid: PERM_CAMSTAB, name: "CAMSTAB"},
	{permid: PERM_NAV_RTH, name: "NAV RTH"},
	{permid: PERM_NAV_POSHOLD, name: "NAV POSHOLD"},
	{permid: PERM_MANUAL, name: "MANUAL"},
	{permid: PERM_BEEPER, name: "BEEPER"},
	{permid: PERM_LEDS_OFF, name: "LEDS OFF"},
	{permid: PERM_LIGHTS, name: "LIGHTS"},
	{permid: PERM_OSD_OFF, name: "OSD OFF"},
	{permid: PERM_TELEMETRY, name: "TELEMETRY"},
	{permid: PERM_AUTO_TUNE, name: "AUTO TUNE"},
	{permid: PERM_BLACKBOX, name: "BLACKBOX"},
	{permid: PERM_FAILSAFE, name: "FAILSAFE"},
	{permid: PERM_NAV_WP, name: "NAV WP"},
	{permid: PERM_AIR_MODE, name: "AIR MODE"},
	{permid: PERM_HOME_RESET, name: "HOME RESET"},
	{permid: PERM_GCS_NAV, name: "GCS NAV"},
	{permid: PERM_FPV_ANGLE_MIX, name: "FPV ANGLE MIX"},
	{permid: PERM_SURFACE, name: "SURFACE"},
	{permid: PERM_FLAPERON, name: "FLAPERON"},
	{permid: PERM_TURN_ASSIST, name: "TURN ASSIST"},
	{permid: PERM_NAV_LAUNCH, name: "NAV LAUNCH"},
	{permid: PERM_SERVO_AUTOTRIM, name: "SERVO AUTOTRIM"},
	{permid: PERM_CAMERA_CONTROL_1, name: "CAMERA CONTROL 1"},
	{permid: PERM_CAMERA_CONTROL_2, name: "CAMERA CONTROL 2"},
	{permid: PERM_CAMERA_CONTROL_3, name: "CAMERA CONTROL 3"},
	{permid: PERM_OSD_ALT_1, name: "OSD ALT 1"},
	{permid: PERM_OSD_ALT_2, name: "OSD ALT 2"},
	{permid: PERM_OSD_ALT_3, name: "OSD ALT 3"},
	{permid: PERM_NAV_COURSE_HOLD, name: "NAV COURSE HOLD"},
	{permid: PERM_MC_BRAKING, name: "MC BRAKING"},
	{permid: PERM_LOITER_CHANGE, name: "LOITER CHANGE"},
	{permid: PERM_MSP_RC_OVERRIDE, name: "MSP RC OVERRIDE"},
	{permid: PERM_PREARM, name: "PREARM"},
	{permid: PERM_TURTLE, name: "TURTLE"},
	{permid: PERM_NAV_CRUISE, name: "NAV CRUISE"},
	{permid: PERM_AUTO_LEVEL_TRIM, name: "AUTO LEVEL TRIM"},
	{permid: PERM_WP_PLANNER, name: "WP PLANNER"},
	{permid: PERM_SOARING, name: "SOARING"},
	{permid: PERM_MISSION_CHANGE, name: "MISSION CHANGE"},
	{permid: PERM_BEEPER_MUTE, name: "BEEPER MUTE"},
	{permid: PERM_MULTI_FUNCTION, name: "MULTI FUNCTION"},
	{permid: PERM_MIXER_PROFILE_2, name: "MIXER PROFILE 2"},
	{permid: PERM_MIXER_TRANSITION, name: "MIXER TRANSITION"},
	{permid: PERM_ANGLE_HOLD, name: "ANGLE HOLD"},
	{permid: PERM_GIMBAL_LEVEL_TILT, name: "GIMBAL LEVEL TILT"},
	{permid: PERM_GIMBAL_LEVEL_ROLL, name: "GIMBAL LEVEL ROLL"},
	{permid: PERM_GIMBAL_CENTER, name: "GIMBAL CENTRE"},
	{permid: PERM_GIMBAL_HEADTRACKER, name: "GIMBAL HEAD TRACKER"},
}

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
	{types.FM_CRUISE3D, []uint16{PERM_NAV_COURSE_HOLD}, "Cruise3D"},
	{types.FM_CRUISE2D, []uint16{PERM_NAV_COURSE_HOLD}, "Cruise2D"},
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
	return 0, "unknown"
}

func perm2name(permid uint8) string {
	for _, p := range pnames {
		if p.permid == permid {
			return p.name
		}
	}
	return "unknown"
}
