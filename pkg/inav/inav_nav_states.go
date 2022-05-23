package inav

func contains(arry []int, key int) bool {
	for _, v := range arry {
		if key == v {
			return true
		}
	}
	return false
}

func IsCruise2d(vers, val int) bool {
	if vers > 0x1ffff { // For 2.0.0, hex = 0x20000
		return contains([]int{29, 30, 31}, val)
	}
	return false
}

func IsCruise3d(vers, val int) bool {
	if vers > 0x1ffff { // For 2.0.0, hex = 0x20000
		return contains([]int{32, 33, 34}, val)
	}
	return false
}

func IsRTH(vers, val int) bool {
	switch {
	case vers > 0x206ff: // 2.7.0 and later
		// 38 is 5.0, BACKTRACK. Should not exist earlier
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 36, 38}, val)
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{8, 9, 10, 11, 12, 13, 14}, val)
	case vers > 0x105ff:
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, val)
	default:
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, val)
	}
}

func IsWP(vers, val int) bool {
	switch {
	case vers > 0x206ff: // 2.7.0 and later
		return contains([]int{15, 16, 17, 18, 19, 20, 21, 35, 37}, val)
	case vers > 0x204ff: // 2.5.0 and later
		return contains([]int{15, 16, 17, 18, 19, 20, 21, 35}, val)
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{15, 16, 17, 18, 19, 20, 21}, val)
	case vers > 0x105ff:
		return contains([]int{20, 21, 22, 23, 24, 25, 26}, val)
	case vers > 0x101ff:
		return contains([]int{22, 23, 24, 25, 26, 27, 28}, val)
	default:
		return contains([]int{22, 23, 24, 25, 26}, val)
	}
}

func IsLaunch(vers, val int) bool {
	switch {
	case vers > 0x1ffff: // 2.0.0 and later
		return contains([]int{25, 26, 28}, val)
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{25, 26, 27, 28}, val)
	case vers > 0x105ff: // For 1.6.0, hex = 0x105ff
		return contains([]int{30, 31, 32, 33}, val)
	case vers > 0x103ff:
		return contains([]int{32, 33, 34, 35}, val)
	}
	return false
}

func IsPH(vers, val int) bool {
	switch {
	case vers > 0x1ffff: // 2.0.0 and later
		return contains([]int{6, 7}, val)
	default:
		return contains([]int{4, 5, 6, 7}, val)
	}
}

func IsAH(vers, val int) bool {
	return contains([]int{2, 3}, val)
}

func IsEmerg(vers, val int) bool {
	switch {
	case vers > 0x10601: // For 1.6.2
		return contains([]int{22, 23, 24}, val)
	case vers > 0x105ff: // 1.6.0
		return contains([]int{27, 28, 29}, val)
	case vers > 0x101ff: // For 1.2.0
		return contains([]int{29, 30, 31}, val)
	default:
		return contains([]int{27, 28, 29}, val)
	}
}

func is_rth_start(vers, val int) bool {
	return val == 8
}

func is_start_land(vers, val int) bool {
	switch {
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{11, 22}, val)
	case vers > 0x105ff: // 1.6 and later
		return contains([]int{16, 27}, val)
	case vers > 0x101ff: // 1.2 and later
		return contains([]int{18, 29}, val)
	default: // prior to 1.2
		return contains([]int{18, 27}, val)
	}
}

func is_landing(vers, val int) bool {
	switch {
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{12, 21, 23}, val)
	case vers > 0x105ff: // 1.6 and later
		return contains([]int{17, 26, 28}, val)
	case vers > 0x101ff: // 1.2 and later
		return contains([]int{19, 28, 30}, val)
	default: // prior to 1.2
		return contains([]int{19, 28}, val)
	}
}

func is_landed(vers, val int) bool {
	switch {
	case vers > 0x10601: // For 1.6.2
		return val == 24
	case vers > 0x105ff: // 1.6.0
		return val == 29
	case vers > 0x101ff: // For 1.2.0
		return val == 31
	default:
		return val == 29
	}

}

func is_hover(vers, val int) bool {
	switch {
	case vers > 0x206ff: // 2.7.0 and later
		return contains([]int{11, 36, 37}, val)
	case vers > 0x10601: // For 1.6.2
		return val == 11
	case vers > 0x105ff: // 1.6.0
		return val == 16
	default:
		return val == 18
	}
}

func NavMode(vers, val int) byte {
	if is_rth_start(vers, val) {
		return 1
	} else if IsPH(vers, val) {
		return 3
	} else if IsWP(vers, val) {
		return 5
	} else if is_start_land(vers, val) {
		return 8
	} else if is_landing(vers, val) {
		return 9
	} else if is_landed(vers, val) {
		return 10
	} else if is_hover(vers, val) {
		return 13
	} else if IsEmerg(vers, val) {
		return 14
	} else {
		return 0
	}
}
