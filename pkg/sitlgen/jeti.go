package sitlgen

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

type JetiChan struct {
	conn net.Conn
}

func NewJetiTX(remote string) (*JetiChan, error) {
	var conn net.Conn
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err == nil {
		conn, err = net.DialTCP("tcp", nil, addr)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Connect JETI to %s\n", remote)
	return &JetiChan{conn: conn}, nil
}

func crc_ccitt_update(crc uint16, v byte) uint16 {
	retval := uint16(0)
	d := v ^ (byte(crc) & 0xff)
	d ^= (d << 4)
	retval = ((uint16(d) << 8) | ((crc & 0xff00) >> 8)) ^ (uint16(d>>4) & 0xff) ^ (uint16(d) << 3)
	return retval
}

func (j *JetiChan) compute_crcz(p []byte) uint16 {
	crc16 := uint16(0)
	for _, v := range p {
		crc16 = crc_ccitt_update(crc16, v)
	}
	return crc16
}

func (j *JetiChan) generate_payload(chans MSPChans, nchan int) []byte {
	psize := byte(2*nchan + 8)
	p := make([]byte, psize)
	p[0] = 0x3e
	p[1] = 3
	p[2] = psize
	p[3] = 6
	p[4] = 0x31
	p[5] = byte(2 * nchan)
	n := 6
	for j := 0; j < nchan; j++ {
		cv := chans[j] * 8
		binary.LittleEndian.PutUint16(p[n:n+2], cv)
		n += 2
	}
	p[n] = 0
	p[n+1] = 0
	return p
}

func (j *JetiChan) Send_TX(chans MSPChans, nchan int) time.Time {
	payload := j.generate_payload(chans, nchan)
	l := len(payload)
	crc := j.compute_crcz(payload[0 : l-2])
	payload[l-2] = byte(crc & 0xff)
	payload[l-1] = byte(crc >> 8)
	//	log.Printf("jeti: %d %+v\n", l, payload)
	_, err := j.conn.Write(payload)
	if err == nil {
		if payload[1] == 3 {
			jp1 := []byte{0x3D, 0x01, 0x08, 0x06, 0x3A, 0x00, 0x98, 0x81}
			_, err = j.conn.Write(jp1)
		}
	}
	return time.Now()
}

func (j *JetiChan) Telem_reader() {
	GenericTelemReader("jeti", j.conn)
}
