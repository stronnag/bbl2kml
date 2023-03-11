package sitlgen

import (
	"encoding/binary"
	"fmt"
	options "github.com/stronnag/bbl2kml/pkg/options"
	"log"
	"net"
	"sort"
	"time"
)

const (
	msp_API_VERSION = 1
	msp_FC_VARIANT  = 2
	msp_FC_VERSION  = 3
	msp_BOARD_INFO  = 4
	msp_BUILD_INFO  = 5

	msp_NAME          = 10
	msp_STATUS        = 101
	msp_SET_RAW_RC    = 200
	msp_RC            = 105
	msp_STATUS_EX     = 150
	msp_RX_MAP        = 64
	msp_MODE_RANGES   = 34
	msp_SET_TX_INFO   = 186
	msp_SET_RX_CONFIG = 45

	msp_COMMON_SETTING = 0x1003
	msp2_INAV_STATUS   = 0x2000
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

const (
	RX_STANDBY = iota
	RX_ARMING
	RX_ARMED
	RX_DISARM
)

const (
	SEND_NONE = iota
	SEND_MSP
	SEND_RSSI
	SEND_CHANS
	SEND_TIMEOUT
)

const SETTING_STR string = "nav_extra_arming_safety"
const MAX_MODE_ACTIVATION_CONDITION_COUNT int = 40

type SChan struct {
	len  uint16
	cmd  uint16
	ok   bool
	data []byte
}

type MSPSerial struct {
	conn    net.Conn
	vcapi   uint16
	fcvers  uint32
	a       uint8
	e       uint8
	r       uint8
	t       uint8
	bypass  bool
	c0      chan SChan
	mranges []ModeRange
	ok      bool
}

type ModeRange struct {
	boxid   byte
	chanidx byte
	start   byte
	end     byte
}

type MSPChans [16]uint16

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
	var paylen uint16
	if len(payload) > 0 {
		paylen = uint16(len(payload))
	}
	buf := make([]byte, 9+paylen)
	buf[0] = '$'
	buf[1] = 'X'
	buf[2] = '<'
	buf[3] = 0 // flags
	binary.LittleEndian.PutUint16(buf[4:6], cmd)
	binary.LittleEndian.PutUint16(buf[6:8], paylen)
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

func (m *MSPSerial) Read_msp(c0 chan SChan) {
	inp := make([]byte, 128)
	var sc SChan
	var count = uint16(0)
	var crc = byte(0)

	n := state_INIT

	for {
		nb, err := m.conn.Read(inp)
		if err == nil && nb > 0 {
			for i := 0; i < nb; i++ {
				switch n {
				case state_INIT:
					if inp[i] == '$' {
						n = state_M
						sc.ok = false
						sc.len = 0
						sc.cmd = 0
					}
				case state_M:
					if inp[i] == 'M' {
						n = state_DIRN
					} else if inp[i] == 'X' {
						n = state_X_HEADER2
					} else {
						n = state_INIT
					}
				case state_DIRN:
					if inp[i] == '!' {
						n = state_LEN
					} else if inp[i] == '>' {
						n = state_LEN
						sc.ok = true
					} else {
						n = state_INIT
					}

				case state_X_HEADER2:
					if inp[i] == '!' {
						n = state_X_FLAGS
					} else if inp[i] == '>' {
						n = state_X_FLAGS
						sc.ok = true
					} else {
						n = state_INIT
					}

				case state_X_FLAGS:
					crc = crc8_dvb_s2(0, inp[i])
					n = state_X_ID1

				case state_X_ID1:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.cmd = uint16(inp[i])
					n = state_X_ID2

				case state_X_ID2:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.cmd |= (uint16(inp[i]) << 8)
					n = state_X_LEN1

				case state_X_LEN1:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.len = uint16(inp[i])
					n = state_X_LEN2

				case state_X_LEN2:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.len |= (uint16(inp[i]) << 8)
					if sc.len > 0 {
						n = state_X_DATA
						count = 0
						sc.data = make([]byte, sc.len)
					} else {
						n = state_X_CHECKSUM
					}
				case state_X_DATA:
					crc = crc8_dvb_s2(crc, inp[i])
					sc.data[count] = inp[i]
					count++
					if count == sc.len {
						n = state_X_CHECKSUM
					}

				case state_X_CHECKSUM:
					ccrc := inp[i]
					if crc != ccrc {
						log.Printf("CRC error on %d\n", sc.cmd)
					} else {
						c0 <- sc
					}
					n = state_INIT

				case state_LEN:
					sc.len = uint16(inp[i])
					crc = inp[i]
					n = state_CMD
				case state_CMD:
					sc.cmd = uint16(inp[i])
					crc ^= inp[i]
					if sc.len == 0 {
						n = state_CRC
					} else {
						sc.data = make([]byte, sc.len)
						n = state_DATA
						count = 0
					}
				case state_DATA:
					sc.data[count] = inp[i]
					crc ^= inp[i]
					count++
					if count == sc.len {
						n = state_CRC
					}
				case state_CRC:
					ccrc := inp[i]
					if crc != ccrc {
						log.Printf("CRC error on %d\n", sc.cmd)
					} else {
						//						log.Fprintf(os.Stderr, "Cmd %v Len %v\n", sc.cmd, sc.len)
						c0 <- sc
					}
					n = state_INIT
				}
			}
		} else {
			m.ok = false
			if err != nil {
				if options.Config.Verbose > 1 {
					log.Printf("Serial Read %v\n", err)
				}
			} else {
				log.Println("serial EOF")
			}

			c0 <- SChan{ok: false}
			m.conn.Close()
			return
		}
	}
}

func NewMSPSerial(remote string) (*MSPSerial, error) {
	var conn net.Conn
	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err == nil {
		conn, err = net.DialTCP("tcp", nil, addr)
	}
	if err != nil {
		return nil, err
	}
	return &MSPSerial{conn: conn, ok: true}, nil
}

func (m *MSPSerial) Send_msp(cmd uint16, payload []byte) {
	if m.ok {
		buf := encode_msp2(cmd, payload)
		_, err := m.conn.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

func (m *MSPSerial) init(nchan chan MSPChans, schan chan byte, rssich chan byte) {
	var fw, api, vers, board, gitrev string
	var v6 bool

	m.c0 = make(chan SChan)
	go m.Read_msp(m.c0)

	m.Send_msp(msp_API_VERSION, nil)
	for done := false; !done; {
		select {
		case v := <-m.c0:
			switch v.cmd {
			case msp_API_VERSION:
				if v.len > 2 {
					api = fmt.Sprintf("%d.%d", v.data[1], v.data[2])
					m.vcapi = uint16(v.data[1])<<8 | uint16(v.data[2])
					m.Send_msp(msp_FC_VARIANT, nil)
				}
			case msp_FC_VARIANT:
				fw = string(v.data[0:4])
				m.Send_msp(msp_FC_VERSION, nil)
			case msp_FC_VERSION:
				vers = fmt.Sprintf("%d.%d.%d", v.data[0], v.data[1], v.data[2])
				m.fcvers = uint32(v.data[0])<<16 | uint32(v.data[1])<<8 | uint32(v.data[2])
				m.Send_msp(msp_BUILD_INFO, nil)
				v6 = (v.data[0] == 6)
			case msp_BUILD_INFO:
				gitrev = string(v.data[19:])
				m.Send_msp(msp_BOARD_INFO, nil)
			case msp_BOARD_INFO:
				if v.len > 8 {
					board = string(v.data[9:])
				} else {
					board = string(v.data[0:4])
				}
				log.Printf("%s v%s %s (%s) API %s\n", fw, vers, board, gitrev, api)
				lstr := len(SETTING_STR)
				buf := make([]byte, lstr+1)
				copy(buf, SETTING_STR)
				buf[lstr] = 0
				m.Send_msp(msp_COMMON_SETTING, buf)

			case msp_COMMON_SETTING:
				if v.len > 0 {
					bystr := v.data[0]
					if v6 {
						bystr++
					}
					if bystr == 2 {
						m.bypass = true
					}
					if options.Config.Verbose > 0 {
						log.Printf("%s: %d (bypass %v)\n", SETTING_STR, bystr, m.bypass)
					}
				}
				m.Send_msp(msp_RX_MAP, nil)

			case msp_RX_MAP:
				if v.len == 4 {
					m.a = uint8(v.data[0]) * 2
					m.e = uint8(v.data[1]) * 2
					m.r = uint8(v.data[2]) * 2
					m.t = uint8(v.data[3]) * 2
					var cmap [4]byte
					cmap[v.data[0]] = 'A'
					cmap[v.data[1]] = 'E'
					cmap[v.data[2]] = 'R'
					cmap[v.data[3]] = 'T'
				} else {
					log.Println("MAP error")
				}
				//				txinfo := []byte{0x02, 0x6c, 0x07, 0xdc, 0x05, 0x4c, 0x04, 0x00, 0x75, 0x03, 0x43, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
				//				m.Send_msp(msp_SET_RX_CONFIG, txinfo)
				//			case msp_SET_RX_CONFIG:
				m.Send_msp(msp_NAME, nil)
			case msp_NAME:
				if v.len > 0 {
					log.Printf("\"%s\"\n", v.data)
				}
				m.Send_msp(msp_MODE_RANGES, nil)
			case msp_MODE_RANGES:
				if v.len > 0 {
					m.deserialise_modes(v.data)
				}
				done = true
			case msp_SET_TX_INFO:
				// RSSI set
			default:
				log.Printf("MSP Unsolicited %d, length %d\n", v.cmd, v.len)
			}
		}
	}
	if options.Config.Verbose > 0 {
		log.Println("Serial init completed, with modes")
		if options.Config.Verbose > 1 {
			for _, u := range m.mranges {
				_, s := mode_to_fm(uint16(u.boxid))
				log.Printf("    %+v %s\n", u, s)
			}
		}
	}
	m.run(nchan, schan, rssich)
}

func (m *MSPSerial) get_ranges() []ModeRange {
	return m.mranges
}

func (m *MSPSerial) deserialise_modes(buf []byte) {
	i := 0
	for j := 0; j < MAX_MODE_ACTIVATION_CONDITION_COUNT; j++ {
		if buf[i+2] != 0 && buf[i+3] != 0 {
			m.mranges = append(m.mranges, ModeRange{buf[i], buf[i+1], buf[i+2], buf[i+3]})
		}
		i += 4
	}
	sort.Slice(m.mranges, func(i, j int) bool {
		if m.mranges[i].chanidx != m.mranges[j].chanidx {
			return m.mranges[i].chanidx < m.mranges[j].chanidx
		}
		return m.mranges[i].start < m.mranges[j].start
	})
}

func (m *MSPSerial) serialise_rx(ichan MSPChans) []byte {
	is := len(ichan)
	buf := make([]byte, is*2)
	for i := 0; i < is; i++ {
		binary.LittleEndian.PutUint16(buf[i*2:2+i*2], ichan[i])
	}
	return buf
}

func deserialise_rx(b []byte) []uint16 {
	bl := binary.Size(b) / 2
	buf := make([]uint16, bl)
	for j := 0; j < bl; j++ {
		n := j * 2
		buf[j] = binary.LittleEndian.Uint16(b[n : n+2])
	}
	return buf
}

func (m *MSPSerial) Rssi(r byte) {
	rk := uint16(r) * 255 / 100
	ra := []byte{byte(rk)}
	m.Send_msp(msp_SET_TX_INFO, ra)
}

func (m *MSPSerial) run(nchan chan MSPChans, schan chan byte, rssich chan byte) {
	ichan := MSPChans{}
	xstatus := uint64(0)
	rssi := byte(0)
	lrssi := byte(0)
	ichan[0] = 1500
	ichan[1] = 1500
	if m.bypass {
		ichan[2] = 1999
	} else {
		ichan[2] = 1500
	}
	ichan[3] = 999
	for j := 4; j < 16; j++ {
		ichan[j] = 1001
	}

	rvstat := byte(0)
	mcnt := byte(0)
	var sv SChan
	tdata := m.serialise_rx(ichan)
	m.Send_msp(msp_SET_RAW_RC, tdata)
	log.Printf("RC init done\n")
	for {
		event := SEND_NONE
		select {
		case v := <-m.c0:
			//			sv_cmd = v.cmd
			//sv_ok = v.ok
			sv = v
			mcnt += 1
			event = SEND_MSP
		case v := <-rssich:
			event = SEND_RSSI
			rssi = v
		case <-time.After(50 * time.Millisecond):
			event = SEND_TIMEOUT
		case v := <-nchan:
			event = SEND_CHANS
			for j, u := range v {
				if u != 0xffff {
					ichan[j] = u
				}
			}
		}
		switch event {
		case SEND_TIMEOUT:
			tdata := m.serialise_rx(ichan)
			m.Send_msp(msp_SET_RAW_RC, tdata)
		case SEND_MSP:
			if sv.ok {
				send_chans := true
				if sv.cmd == msp_SET_RAW_RC && mcnt%5 == 0 {
					send_chans = false
					m.Send_msp(msp2_INAV_STATUS, nil)
				} else if sv.cmd == msp2_INAV_STATUS {
					status := binary.LittleEndian.Uint64(sv.data[13:21])
					armflags := binary.LittleEndian.Uint32(sv.data[9:13])
					if options.Config.Verbose > 2 {
						log.Printf("Status, Armflags  %x %x\n", status, armflags)
					}
					// Unarmed, able to arm
					if (status&1) == 0 && armflags < 0x200 {
						if rvstat != 1 {
							rvstat = 1
							if options.Config.Verbose > 1 {
								log.Printf("Set status %d (%x)\n", rvstat, armflags)
							}
							schan <- 1
						}
					}
					if (xstatus & 1) != (status & 1) {
						if options.Config.Verbose > 0 {
							log.Printf("status changed %x -> %x (%x)\n", xstatus, status, armflags)
						}
						xstatus = status
						if (status & 1) == 0 {
							if rvstat != 2 {
								rvstat = 2
								if options.Config.Verbose > 1 {
									log.Printf("Set status %d (%x)\n", rvstat, armflags)
								}
								schan <- 2
							}
						} else {
							if rvstat != 3 {
								rvstat = 3
								if options.Config.Verbose > 1 {
									log.Printf("Set status %d (%x)\n", rvstat, armflags)
								}
								schan <- 3
							}
						}
					}
				}

				if send_chans {
					tdata := m.serialise_rx(ichan)
					m.Send_msp(msp_SET_RAW_RC, tdata)
				}
			} else {
				if options.Config.Verbose > 1 {
					log.Println("RC data send failed")
				}
				schan <- 0xff
			}

		case SEND_RSSI:
			if lrssi != rssi {
				m.Rssi(rssi)
				lrssi = rssi
			}
		case SEND_CHANS:
			tdata := m.serialise_rx(ichan)
			m.Send_msp(msp_SET_RAW_RC, tdata)
		}
	}
}

func (m *MSPSerial) Close() {
	if m.ok {
		m.ok = false
		m.conn.Close()
	}
}
