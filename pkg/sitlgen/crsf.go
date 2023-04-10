package sitlgen

import (
	"log"
	"net"
	"time"
)

type CrsfChan struct {
	conn net.Conn
}

func (c *CrsfChan) pack_chans(ichans MSPChans, buf []byte) {
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

func NewCrsfTX(remote string) (*CrsfChan, error) {
	var conn net.Conn
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err == nil {
		conn, err = net.DialTCP("tcp", nil, addr)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Connect CRSF to %s\n", remote)
	return &CrsfChan{conn: conn}, nil
}

func (c *CrsfChan) Send_TX(chans MSPChans, nchan int) time.Time {
	buf := make([]byte, 26)
	buf[0] = 0xc8
	buf[1] = 24
	buf[2] = 0x16
	c.pack_chans(chans, buf[3:])
	crc := byte(0)
	for _, b := range buf[2:25] {
		crc = rx_crc8_dvb_s2(crc, b)
	}
	buf[25] = crc
	c.conn.Write(buf)
	return time.Now()
}

func (c *CrsfChan) Telem_reader() {
	GenericTelemReader("crsf", c.conn)
}
