package sitlgen

import (
	"log"
	"net"
	"time"
)

type SbusChan struct {
	conn net.Conn
}

func (s *SbusChan) pack_chans(ichans MSPChans, buf []byte) {
	var chans [16]uint16
	for j := 0; j < 16; j++ {
		chans[j] = (ichans[j] - 880) * 8 / 5
	}

	buf[0] = byte(chans[0] & rx_mask(8))
	buf[1] = byte(chans[0] >> 8)
	buf[1] |= byte(chans[1]&rx_mask(5)) << 3
	buf[2] = byte(chans[1] >> 5)
	buf[2] |= byte(chans[2]&rx_mask(2)) << 6
	buf[3] = byte(chans[2] >> 2)
	buf[4] = byte(chans[2] >> 10)
	buf[4] |= byte(chans[3] & rx_mask(7) << 1)
	buf[5] = byte(chans[3] >> 7)
	buf[5] |= byte(chans[4] & rx_mask(4) << 4)
	buf[6] = byte(chans[4] >> 4)
	buf[6] |= byte(chans[5] & rx_mask(1) << 7)
	buf[7] = byte(chans[5] >> 1)
	buf[8] = byte(chans[5] >> 9)
	buf[8] |= byte(chans[6] & rx_mask(6) << 2)
	buf[9] = byte(chans[6] >> 6)
	buf[9] |= byte(chans[7] & rx_mask(3) << 5)
	buf[10] = byte(chans[7] >> 3)

	buf[11] = byte(chans[8] & rx_mask(8))
	buf[12] = byte(chans[8] >> 8)
	buf[12] |= byte(chans[9]&rx_mask(5)) << 3
	buf[13] = byte(chans[9] >> 5)
	buf[13] |= byte(chans[10]&rx_mask(2)) << 6
	buf[14] = byte(chans[10] >> 2)
	buf[15] = byte(chans[10] >> 10)
	buf[15] |= byte(chans[11] & rx_mask(7) << 1)
	buf[16] = byte(chans[11] >> 7)
	buf[16] |= byte(chans[12] & rx_mask(4) << 4)
	buf[17] = byte(chans[12] >> 4)
	buf[17] |= byte(chans[13] & rx_mask(1) << 7)
	buf[18] = byte(chans[13] >> 1)
	buf[19] = byte(chans[13] >> 9)
	buf[19] |= byte(chans[14] & rx_mask(6) << 2)
	buf[20] = byte(chans[14] >> 6)
	buf[20] |= byte(chans[15] & rx_mask(3) << 5)
	buf[21] = byte(chans[15] >> 3)
}

func NewSbusTX(remote string) (*SbusChan, error) {
	var conn net.Conn
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err == nil {
		conn, err = net.DialTCP("tcp", nil, addr)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Connect SBUS to %s\n", remote)
	return &SbusChan{conn: conn}, nil
}

func (s *SbusChan) Send_TX(chans MSPChans, nchan int) time.Time {
	sbuf := make([]byte, 25) // SbusFrame(SbusFrame{syncb: 0x1f, endb: 0, flag: 3})
	sbuf[0] = 0xf
	s.pack_chans(chans, sbuf[1:])
	sbuf[23] = 3
	sbuf[24] = 0
	s.conn.Write(sbuf)
	return time.Now()
}

func (s *SbusChan) Telem_reader() {
	GenericTelemReader("sbus", s.conn)
}
