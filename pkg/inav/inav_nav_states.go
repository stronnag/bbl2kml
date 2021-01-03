package inav

func contains(arry []int, key int) bool {
	for _, v := range arry {
		if key == v {
			return true
		}
	}
	return false
}

func IsCruise2d(val, vers int) bool {
	if vers > 0x1ffff { // For 2.0.0, hex = 0x20000
		return contains([]int{29, 30, 31}, val)
	}
	return false
}

func IsCruise3d(val, vers int) bool {
	if vers > 0x1ffff { // For 2.0.0, hex = 0x20000
		return contains([]int{32, 33, 34}, val)
	}
	return false
}

func IsRTH(vers, val int) bool {
	switch {
	case vers > 0x206ff: // 2.7.0 and later
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 36}, val)
	case vers > 0x10601: // For 1.6.2, hex = 0x10601
		return contains([]int{8, 9, 10, 11, 12, 13, 14}, val)
	case vers > 0x105ff:
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}, val)
	default:
		return contains([]int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, val)
	}
	return false
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
	return false
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
