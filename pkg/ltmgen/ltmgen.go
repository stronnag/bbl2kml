package ltmgen

import (
	"strings"
	"strconv"
	"fmt"
	"encoding/binary"
	"time"
	"os"
	"log"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
	inav "github.com/stronnag/bbl2kml/pkg/inav"
)

type ltmbuf struct {
	msg []byte
	len byte
}

func newLTM(mtype byte) *ltmbuf {
	paylen := byte(0)
	switch mtype {
	case 'A':
		paylen = 6
	case 'G':
		paylen = 14
	case 'N':
		paylen = 6
	case 'O':
		paylen = 14
	case 'S':
		paylen = 7
	case 'X':
		paylen = 6
	case 'x':
		paylen = 1
	case 'q':
		paylen = 2
	default:
		log.Fatal("LTM: No payload defined for type '%c'\n", mtype)
	}

	buf := make([]byte, paylen+4)
	buf[0] = '$'
	buf[1] = 'T'
	buf[2] = mtype
	ltm := &ltmbuf{buf, paylen}
	return ltm
}

func (l *ltmbuf) String() string {
	var sb strings.Builder
	for _, s := range l.msg {
		fmt.Fprintf(&sb, "%02x ", s)
	}
	return strings.TrimSpace(sb.String())
}
func (l *ltmbuf) checksum() {
	c := byte(0)
	for _, s := range l.msg[3:] {
		c = c ^ s
	}
	l.msg[l.len+3] = c
}

func (l *ltmbuf) aframe(b types.LogItem) {
	binary.LittleEndian.PutUint16(l.msg[3:5], uint16(b.Pitch))
	binary.LittleEndian.PutUint16(l.msg[5:7], uint16(b.Roll))
	binary.LittleEndian.PutUint16(l.msg[7:9], uint16(b.Cse))
	l.checksum()
}

func (l *ltmbuf) gframe(b types.LogItem) {
	lat := int32(b.Lat * 1.0e7)
	lon := int32(b.Lon * 1.0e7)
	alt := int32(b.Alt * 100)
	binary.LittleEndian.PutUint32(l.msg[3:7], uint32(lat))
	binary.LittleEndian.PutUint32(l.msg[7:11], uint32(lon))
	l.msg[11] = byte(b.Spd)
	binary.LittleEndian.PutUint32(l.msg[12:16], uint32(alt))
	l.msg[16] = b.Fix | (b.Numsat << 2)
	l.checksum()
}

func (l *ltmbuf) nframe(b types.LogItem, action byte, wpno byte) {
	switch b.Fmode {
	case types.FM_AH, types.FM_PH:
		l.msg[3] = 1
	case types.FM_RTH:
		l.msg[3] = 2
	case types.FM_WP:
		l.msg[3] = 3
	default:
		l.msg[3] = 0
	}
	if b.NavState != -1 {
		l.msg[4] = byte(b.NavState)
	} else {
		l.msg[4] = 0 // synthesise
	}
	l.msg[5] = action
	l.msg[6] = wpno
	l.msg[7] = 0
	l.msg[8] = 0
	l.checksum()
}

func (l *ltmbuf) oframe(b types.LogItem, hlat float64, hlon float64) {
	lat := int32(hlat * 1.0e7)
	lon := int32(hlon * 1.0e7)
	binary.LittleEndian.PutUint32(l.msg[3:7], uint32(lat))
	binary.LittleEndian.PutUint32(l.msg[7:11], uint32(lon))
	binary.LittleEndian.PutUint32(l.msg[11:15], 0)
	l.msg[15] = 1
	l.msg[16] = b.Fix
	l.checksum()
}

func ltm_flight_mode(fm uint8) uint8 {
	var fms byte
	switch fm {
	case types.FM_ACRO:
		fms = 1
	case types.FM_MANUAL:
		fms = 0
	case types.FM_HORIZON:
		fms = 3
	case types.FM_ANGLE:
		fms = 2
	case types.FM_LAUNCH:
		fms = 20
	case types.FM_RTH:
		fms = 13
	case types.FM_WP:
		fms = 10
	case types.FM_CRUISE3D, types.FM_CRUISE2D:
		fms = 18
	case types.FM_PH:
		fms = 9
	case types.FM_AH:
		fms = 8
	default:
		fms = 0
	}
	return (fms << 2)
}

func (l *ltmbuf) sframe(b types.LogItem) {
	binary.LittleEndian.PutUint16(l.msg[3:5], uint16(1000*b.Volts)) // units ??
	binary.LittleEndian.PutUint16(l.msg[5:7], uint16(b.Energy))     // units
	l.msg[7] = uint8(255 * int(b.Rssi) / 100)
	l.msg[8] = byte(b.Spd)
	l.msg[9] = (b.Status & (types.Is_ARMED | types.Is_FAIL)) | ltm_flight_mode(b.Fmode)
	l.checksum()
}

func (l *ltmbuf) xframe(hdop uint16, xcount uint8) {
	binary.LittleEndian.PutUint16(l.msg[3:5], hdop)
	l.msg[5] = 0
	l.msg[6] = xcount
	l.msg[7] = 0
	l.checksum()
}

func (l *ltmbuf) lxframe(r byte) {
	l.msg[3] = r
	l.checksum()
}

func (l *ltmbuf) qframe(d uint16) {
	binary.LittleEndian.PutUint16(l.msg[3:5], d)
	l.checksum()
}

func read_mission() *mission.Mission {
	var ms *mission.Mission
	ms = nil
	if len(options.Mission) > 0 {
		var err error
		_, ms, err = mission.Read_Mission_File(options.Mission)
		if err == nil {
			for k, mi := range ms.MissionItems {
				if mi.Is_GeoPoint() && geo.Getfrobnication() {
					ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, _ = geo.Frobnicate_move(ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, 0)
				}
				if mi.Action == "JUMP" {
					ms.MissionItems[k].P3 = ms.MissionItems[k].P2
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "* Failed to read mission file %s\n", options.Mission)
		}
	}
	return ms
}

func LTMGen(seg types.LogSegment, meta types.FlightMeta) {
	verbose := false
	fast := false
	var s *MSPSerial

	typ := byte(0)
	switch meta.Motors {
	case 0, 1, 2:
		typ = 8
	case 3:
		typ = 1
	case 4:
		typ = 3
	case 6:
		typ = 7
	case 8:
		typ = 11
	}

	parts := strings.Split(options.LTMdev, ",")
	switch len(parts) {
	case 3:
		fast = true
		fallthrough
	case 2:
		if len(parts[1]) == 1 {
			t, _ := strconv.ParseInt(parts[1], 10, 8)
			typ = byte(t)
		}
		fallthrough
	case 1:
		s = NewMSPSerial(parts[0], 0)
	default:
		return
	}

	laststat := uint8(255)
	nvs := 0
	tgt := 0

	xcount := uint8(0)
	ld := uint16(0)

	var st, lt time.Time
	var hlon, hlat float64

	if seg.H.Flags&types.HOME_SAFE != 0 {
		hlat = seg.H.SafeLat
		hlon = seg.H.SafeLon
	} else {
		hlat = seg.H.HomeLat
		hlon = seg.H.HomeLon
	}

	ms := read_mission()

	if meta.Flags&types.Has_Firmware != 0 {
		s.Write(MSP_serialise_ident(typ))
		s.Write(MSP_serialise_api_version())
		parts := strings.Split(meta.Firmware, " ")
		lp := len(parts)
		if lp > 3 {
			s.Write(MSP_serialise_board_info(parts[3]))
		}
		if lp > 0 {
			s.Write(MSP_serialise_fc_variant(parts[0]))
			if lp > 1 {
				vers := make([]byte, 3)
				vers[0] = parts[1][0] - '0'
				vers[1] = parts[1][2] - '0'
				vers[2] = parts[1][4] - '0'
				s.Write(MSP_serialise_fc_version(vers))
				if lp > 2 {
					s.Write(MSP_serialise_build_info(parts[2][1 : len(parts[2])-2]))
				}
			}
		}
	}

	if meta.Flags&types.Has_Craft != 0 {
		s.Write(MSP_serialise_name(meta.Craft))
	}

	if meta.Sensors != 0 {
		s.Write(MSP_serialise_status(meta.Sensors))
	}

	for _, b := range seg.L.Items {

		if st.IsZero() {
			st = b.Utc
		}

		if b.Fmode != laststat {
			switch b.Fmode {
			case types.FM_WP:
				if ms != nil {
					tgt = 1
					nvs = 5
				}
			case types.FM_RTH:
				tgt = 0
				nvs = 1
			case types.FM_PH:
				tgt = 0
				nvs = 3
			default:
				tgt = 0
				nvs = 0
			}
			if b.NavState == -1 {
				b.NavState = nvs
			}
			l := newLTM('N')
			l.nframe(b, 0, 0)
			s.Write(l.msg)
			laststat = b.Fmode
		}

		if b.Fmode == types.FM_WP && ms != nil {
			act := 0
			tgt, nvs, act = inav.WP_state(ms, b, tgt, nvs)
			b.NavState = nvs
			l := newLTM('N')
			l.nframe(b, byte(act), byte(tgt))
			s.Write(l.msg)
		}

		tdiff := b.Utc.Sub(lt)

		l := newLTM('G')
		l.gframe(b)
		s.Write(l.msg)
		if verbose {
			fmt.Fprintf(os.Stderr, "Gframe : %s\n", l)
		}

		l = newLTM('A')
		l.aframe(b)
		s.Write(l.msg)
		if verbose {
			fmt.Fprintf(os.Stderr, "Aframe : %s\n", l)
		}

		l = newLTM('O')
		l.oframe(b, hlat, hlon)
		s.Write(l.msg)
		if verbose {
			fmt.Fprintf(os.Stderr, "Oframe : %s\n", l)
		}

		l = newLTM('S')
		l.sframe(b)
		s.Write(l.msg)
		if verbose {
			fmt.Fprintf(os.Stderr, "Sframe : %s\n", l)
		}

		l = newLTM('X')
		l.xframe(b.Hdop, xcount)
		s.Write(l.msg)
		xcount = (xcount + 1) & 0xff
		if verbose {
			fmt.Fprintf(os.Stderr, "Xframe : %s\n", l)
		}

		// Nframe, may be BBL only

		if !lt.IsZero() {
			if fast {
				time.Sleep(10 * time.Millisecond)
			} else if tdiff > 0 {
				time.Sleep(tdiff)
			}
		}

		et := b.Utc.Sub(st)
		d := uint16(et.Seconds())
		if d != ld {
			l = newLTM('q')
			l.qframe(d)
			s.Write(l.msg)
			ld = d
		}
		lt = b.Utc
	}
	l := newLTM('x')
	l.lxframe(byte(meta.Disarm))
	s.Write(l.msg)
	s.Close()
}
