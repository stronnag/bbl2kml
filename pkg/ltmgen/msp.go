package ltmgen

import (
	"encoding/binary"
)

const (
	msp_API_VERSION = 1
	msp_FC_VARIANT  = 2
	msp_FC_VERSION  = 3
	msp_BOARD_INFO  = 4
	msp_BUILD_INFO  = 5
	msp_NAME        = 10
	msp_IDENT       = 100
	msp_STATUS      = 101
)
const (
	state_INIT = iota
	state_M
	state_DIRN
	state_LEN
	state_CMD
	state_DATA
	state_CRC
	state_X_HEADER2
	state_X_FLAGS
	state_X_ID1
	state_X_ID2
	state_X_LEN1
	state_X_LEN2
	state_X_DATA
	state_X_CHECKSUM
)

func crc8_dvb_s2(crc byte, a byte) byte {
	crc ^= a
	for i := 0; i < 8; i++ {
		if (crc & 0x80) != 0 {
			crc = (crc << 1) ^ 0xd5
		} else {
			crc = crc << 1
		}
	}
	return crc
}

func encode_msp2(cmd uint16, payload []byte) []byte {
	var paylen int16
	if len(payload) > 0 {
		paylen = int16(len(payload))
	}
	buf := make([]byte, 9+paylen)
	buf[0] = '$'
	buf[1] = 'X'
	buf[2] = '>'
	buf[3] = 0 // flags
	binary.LittleEndian.PutUint16(buf[4:6], cmd)
	binary.LittleEndian.PutUint16(buf[6:8], uint16(paylen))
	if paylen > 0 {
		copy(buf[8:], payload)
	}
	crc := byte(0)
	for _, b := range buf[3 : paylen+8] {
		crc = crc8_dvb_s2(crc, b)
	}
	buf[8+paylen] = crc
	return buf
}

func encode_msp1(cmd uint16, payload []byte) []byte {
	var paylen byte
	if len(payload) > 0 {
		paylen = byte(len(payload))
	}
	buf := make([]byte, 6+paylen)
	buf[0] = '$'
	buf[1] = 'M'
	buf[2] = '>'
	buf[3] = paylen
	buf[4] = byte(cmd)
	if paylen > 0 {
		copy(buf[5:], payload)

	}
	crc := byte(0)
	for _, b := range buf[3:] {
		crc ^= b
	}
	buf[5+paylen] = crc
	return buf
}

func encode_msp(cmd uint16, payload []byte) []byte {
	return encode_msp2(cmd, payload)
}

func MSP_serialise_ident(typ byte) []byte {
	buf := make([]byte, 7)
	buf[0] = 0xff
	buf[1] = typ
	buf[2] = 42
	return encode_msp(msp_IDENT, buf)
}

func MSP_serialise_api_version() []byte {
	buf := make([]byte, 3)
	buf[1] = 2
	buf[2] = 0
	return encode_msp(msp_API_VERSION, buf)
}

func MSP_serialise_board_info(fcname string) []byte {
	buf := make([]byte, 9+len(fcname))
	copy(buf[9:], []byte(fcname))
	return encode_msp(msp_BOARD_INFO, buf)
}

func MSP_serialise_name(name string) []byte {
	buf := make([]byte, len(name))
	copy(buf, []byte(name))
	return encode_msp(msp_NAME, buf)
}

func MSP_serialise_fc_variant(name string) []byte {
	buf := make([]byte, len(name))
	copy(buf, []byte(name))
	return encode_msp(msp_FC_VARIANT, buf)
}

func MSP_serialise_fc_version(vers []byte) []byte {
	return encode_msp(msp_FC_VERSION, vers)
}

func MSP_serialise_build_info(gitvers string) []byte {
	buf := make([]byte, 19+len(gitvers))
	copy(buf[19:], []byte(gitvers))
	return encode_msp(msp_BUILD_INFO, buf)
}

func MSP_serialise_status(sensors uint16) []byte {
	buf := make([]byte, 11)
	binary.LittleEndian.PutUint16(buf[4:6], sensors)
	return encode_msp(msp_STATUS, buf)
}
