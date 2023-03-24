package sitlgen

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"time"

	mission "github.com/stronnag/bbl2kml/pkg/mission"
	options "github.com/stronnag/bbl2kml/pkg/options"
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
	msp_BOXNAMES      = 116
	msp_SET_WP        = 209

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
	wp_WAYPOINT = 1 + iota
	wp_POSHOLD_UNLIM
	wp_POSHOLD_TIME
	wp_RTH
	wp_SET_POI
	wp_JUMP
	wp_SET_HEAD
	wp_LAND
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

type MSPChans [18]uint16

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
	inp := make([]byte, 512)
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
func encode_action(a string) byte {
	var b byte
	switch a {
	case "WAYPOINT":
		b = wp_WAYPOINT
	case "POSHOLD_UNLIM":
		b = wp_POSHOLD_UNLIM
	case "POSHOLD_TIME":
		b = wp_POSHOLD_TIME
	case "RTH":
		b = wp_RTH
	case "SET_POI":
		b = wp_SET_POI
	case "JUMP":
		b = wp_JUMP
	case "SET_HEAD":
		b = wp_SET_HEAD
	case "LAND":
		b = wp_LAND
	default:
		b = wp_WAYPOINT
	}
	return b
}

func serialise_wp(mi mission.MissionItem, last bool) []byte {
	buf := make([]byte, 21)
	buf[0] = byte(mi.No)
	buf[1] = encode_action(mi.Action)
	v := int32(mi.Lat * 1e7)
	binary.LittleEndian.PutUint32(buf[2:6], uint32(v))
	v = int32(mi.Lon * 1e7)
	binary.LittleEndian.PutUint32(buf[6:10], uint32(v))
	binary.LittleEndian.PutUint32(buf[10:14], uint32(100*mi.Alt))
	binary.LittleEndian.PutUint16(buf[14:16], uint16(mi.P1))
	binary.LittleEndian.PutUint16(buf[16:18], uint16(mi.P2))
	binary.LittleEndian.PutUint16(buf[18:20], uint16(mi.P3))
	buf[20] = mi.Flag
	return buf
}

func (m *MSPSerial) upload_mission(ms *mission.Mission) {
	mlen := len(ms.MissionItems) - 1
	for j, mi := range ms.MissionItems {
		mi.No = j + 1
		b := serialise_wp(mi, (j == mlen))
		m.Send_msp(msp_SET_WP, b)
		v := <-m.c0
		if !v.ok {
			break
		}
	}
	Sitl_logger(2, "Uploaded mission")
}

func (m *MSPSerial) init(nchan chan RCInfo, schan chan byte, mintime int) {
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
				m.Send_msp(msp_BOXNAMES, nil)
			case msp_BOXNAMES:
				if v.len > 0 {
					Sitl_logger(2, "%s\n", v.data)
				}
				done = true
			case msp_SET_TX_INFO:
				// RSSI set (UNUSED for now)
			default:
				log.Printf("MSP Unsolicited %d, length %d\n", v.cmd, v.len)
			}
		}
	}

	if options.Config.Mission != "" {
		_, mm, _ := mission.Read_Mission_File_Index(options.Config.Mission, options.Config.MissionIndex)
		if mm != nil {
			m.upload_mission(mm)
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
	m.run(nchan, schan, int64(mintime))
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

func (m *MSPSerial) send_tx(ichan MSPChans) time.Time {
	tdata := m.serialise_rx(ichan)
	m.Send_msp(msp_SET_RAW_RC, tdata)
	return time.Now()
}

type StatusInfo struct {
	boxflags uint64
	armflags uint32
	rvstat   byte
}

func (m *MSPSerial) run(nchan chan RCInfo, schan chan byte, mintime int64) {
	ichan := MSPChans{}
	si := StatusInfo{}

	rssi := byte(0)
	lrssi := byte(0)
	xfs := byte(0)

	ichan[0] = 1500
	ichan[1] = 1500
	if m.bypass {
		ichan[2] = 1999
	} else {
		ichan[2] = 1500
	}
	ichan[3] = 999
	for j := 4; j < len(ichan); j++ {
		ichan[j] = 1001
	}

	mcnt := byte(0)

	ntx := 1
	inflight := byte(0)

	inflight |= 1
	stime := m.send_tx(ichan)
	log.Printf("RC init done\n")

	nstat := 0
	nrssi := 0

	start := time.Now()
	last := start

	for {
		select {
		case v := <-m.c0:
			if v.ok {
				mcnt += 1
				switch v.cmd {
				case msp_SET_RAW_RC:
					ntx += 1
					inflight &= ^byte(1)
					if inflight == 0 {
						inflight |= 2
						m.Send_msp(msp2_INAV_STATUS, nil)
					}
				case msp2_INAV_STATUS:
					nstat += 1
					inflight &= ^byte(2)
					si.parse_status(schan, v.data)
					if inflight == 0 {
						if rssi != lrssi {
							lrssi = rssi
							inflight |= 4
							m.Rssi(rssi)
						}
					}
				case msp_SET_TX_INFO:
					nrssi += 1
					inflight &= ^byte(4)
				}
			} else {
				if options.Config.Verbose > 1 {
					log.Println("RC data send failed")
				}
				schan <- 0xff
			}
		case <-time.After(25 * time.Millisecond):

		case v := <-nchan:
			cchan := false
			for j, u := range v.chans {
				if u != 0xffff && u != ichan[j] {
					ichan[j] = u
					if j > 3 || j == 3 && u < 900 {
						cchan = true
					}
				}
			}
			if cchan {
				Sitl_logger(3, "%s\n", dump_channels(ichan))
			}
			xfs = v.fs
			rssi = v.rssi
		}

		if time.Since(stime) > time.Duration(mintime)*time.Millisecond {
			if ichan[3] == 0xd0d0 { // Failsafe, ignore
				stime = time.Now()
				if inflight == 0 {
					inflight |= 2
					m.Send_msp(msp2_INAV_STATUS, nil)
				}
			} else {
				if inflight == 0 {
					inflight |= 1
					stime = m.send_tx(ichan)
				}
			}
		}

		if options.Config.Verbose > 2 {
			now := time.Since(last)
			if now > time.Duration(10*time.Second) {
				d := time.Since(start)
				secs := int(d.Seconds())
				log.Printf("Stats %ds: Tx: %d RSSI %d Stats %d (%d)\n", secs, ntx, nrssi, nstat, xfs)
				last = time.Now()
			}
		}
	}
}

func (s *StatusInfo) parse_status(schan chan byte, data []byte) {
	armflags := binary.LittleEndian.Uint32(data[9:13])
	boxflags := binary.LittleEndian.Uint64(data[13:21])

	if options.Config.Verbose > 2 {
		if !((s.boxflags == boxflags) && (armflags == s.armflags)) {
			log.Printf("Boxflags: %x Armflags: %s\n", boxflags, arm_status(armflags))
		}
	}
	// Unarmed, able to arm
	if (boxflags&1 == 0) && armflags < 0x80 {
		if s.rvstat != 1 {
			s.rvstat = 1
			if options.Config.Verbose > 1 {
				log.Printf("Set status %d (%x)\n", s.rvstat, armflags)
			}
			schan <- 1
		}
	}
	if (s.boxflags & 1) != (boxflags & 1) {
		if options.Config.Verbose > 0 {
			log.Printf("boxflags changed %x -> %x (%x)\n", s.boxflags, boxflags, armflags)
		}
		if (boxflags & 1) == 0 {
			if s.rvstat != 2 {
				s.rvstat = 2
				if options.Config.Verbose > 1 {
					log.Printf("Set boxflags %d (%x)\n", s.rvstat, armflags)
				}
				schan <- 2
			}
		} else {
			if s.rvstat != 3 {
				s.rvstat = 3
				if options.Config.Verbose > 1 {
					log.Printf("Set boxflags %d (%x)\n", s.rvstat, armflags)
				}
				schan <- 3
			}
		}
	}
	s.boxflags = boxflags
	s.armflags = armflags
}

func dump_channels(chans MSPChans) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for j, v := range chans {
		switch j {
		case 0:
			sb.WriteString("A:")
		case 1:
			sb.WriteString("E:")
		case 2:
			sb.WriteString("R:")
		case 3:
			sb.WriteString("T:")
		default:
			fmt.Fprintf(&sb, "%d:", j+1)
		}
		if v == 0xd0d0 {
			sb.WriteString("F/S ")
		} else {
			fmt.Fprintf(&sb, "%d", v)
		}
		if j != len(chans)-1 {
			sb.WriteString(", ")
		} else {
			sb.WriteByte(']')
		}
	}
	return sb.String()
}

func (m *MSPSerial) Close() {
	if m.ok {
		m.ok = false
		m.conn.Close()
	}
}

func arm_status(status uint32) string {
	armfails := [...]string{
		"",           /*      1 */
		"",           /*      2 */
		"Armed",      /*      4 */
		"Ever armed", /*      8 */
		"",           /*     10 */ // HITL
		"",           /*     20 */ // SITL
		"",           /*     40 */
		"F/S",        /*     80 */
		"Level",      /*    100 */
		"Calibrate",  /*    200 */
		"Overload",   /*    400 */
		"NavUnsafe", "MagCal", "AccCal", "ArmSwitch", "H/WFail",
		"BoxF/S", "BoxKill", "RCLink", "Throttle", "CLI",
		"CMS", "OSD", "Roll/Pitch", "Autotrim", "OOM",
		"Settings", "PWM Out", "PreArm", "DSHOTBeep", "Land", "Other",
	}

	var sarry []string
	if status < 0x80 {
		if status&(1<<2) != 0 {
			sarry = append(sarry, armfails[2])
		}
		if len(sarry) == 0 {
			sarry = append(sarry, "Ready to arm")
		}
	} else {
		for i := 0; i < len(armfails); i++ {
			if ((status & (1 << i)) != 0) && armfails[i] != "" {
				sarry = append(sarry, armfails[i])
			}
		}
	}
	sarry = append(sarry, fmt.Sprintf("(0x%x)", status))
	return strings.Join(sarry, " ")
}
