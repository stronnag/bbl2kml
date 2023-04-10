package sitlgen

import (
	"log"
	"net"
	"time"
)

const (
	IBUS_MAX_SLOT   = 14
	IBUS_MAX_CHAN   = 18
	IBUS_CRCBASE    = 0xffff
	IBUS_FRAME_SIZE = 32
)

type IbusChan struct {
	conn net.Conn
}

func (c *IbusChan) pack_chans(chans MSPChans, buf []byte) {
	k := 2
	for j := 0; j < IBUS_MAX_SLOT; j++ {
		buf[k] = byte(chans[j] & 0xff)
		buf[k+1] = byte((chans[j] >> 8) & 0x0f)
		k += 2
	}
	k = 3
	for j := IBUS_MAX_SLOT; j < IBUS_MAX_CHAN; j++ {
		buf[k] |= byte((chans[j] & 0x0f) << 4)
		buf[k+2] |= byte(chans[j] & 0xf0)
		buf[k+4] |= byte(uint16((chans[j] >> 8) << 4))
		k += 6
	}
}

func NewIbusTX(remote string) (*IbusChan, error) {
	var conn net.Conn
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err == nil {
		conn, err = net.DialTCP("tcp", nil, addr)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Connect IBUS to %s\n", remote)
	return &IbusChan{conn: conn}, nil
}

func (c *IbusChan) Send_TX(chans MSPChans, nchan int) time.Time {
	buf := make([]byte, IBUS_FRAME_SIZE)
	buf[0] = IBUS_FRAME_SIZE
	buf[1] = 0x40 // Allegedly
	c.pack_chans(chans, buf)

	crc := uint16(IBUS_CRCBASE)
	for _, b := range buf[0 : IBUS_FRAME_SIZE-2] {
		crc -= uint16(b)
	}
	buf[IBUS_FRAME_SIZE-2] = byte(crc & 0xff)
	buf[IBUS_FRAME_SIZE-1] = byte(crc >> 8)

	c.conn.Write(buf)
	return time.Now()
}

func (c *IbusChan) Telem_reader() {
	GenericTelemReader("ibus", c.conn)
}
