package sitlgen

import "time"

type TxChan interface {
	Send_TX(MSPChans, int) time.Time
	Telem_reader()
}

func rx_crc8_dvb_s2(crc byte, a byte) byte {
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

func rx_mask(n int) uint16 {
	mb := uint16(1<<n) - 1
	return mb
}
