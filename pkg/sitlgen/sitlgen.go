package sitlgen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/eiannone/keyboard"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const MODE_OFFSET = 4

type SimData struct {
	Lat    float32
	Lon    float32
	Alt    float32
	Galt   float32
	Speed  float32
	Cog    float32
	Roll   float32
	Pitch  float32
	Yaw    float32
	Gyro_x float32
	Gyro_y float32
	Gyro_z float32
	Acc_x  float32
	Acc_y  float32
	Acc_z  float32
	RC_a   uint16
	RC_e   uint16
	RC_r   uint16
	RC_t   uint16
	Fmode  uint16
	Rssi   byte
}

type SitlGen struct {
	drefmap map[string]uint32
	mchans  MSPChans
	swchan  int16
	swval   uint16
}

func NewSITL() *SitlGen {
	return &SitlGen{drefmap: make(map[string]uint32), mchans: MSPChans{}, swchan: -1, swval: 0}
}

func setvalue(r ModeRange) uint16 {
	return uint16(r.end+r.start)*25/2 + 900
}

func clrvalue(mr []ModeRange, r ModeRange) uint16 {
	smin := byte(255)
	for _, m := range mr {
		if m.chanidx == r.chanidx {
			if m.start < smin {
				smin = m.start
			}
		}
	}
	if smin == 255 {
		smin = 4
	}
	return uint16(smin-1)*25 + 900 + 10
}

func (x *SitlGen) change_mode(mranges []ModeRange, _from, _to uint16) {
	from, fstr := fm_to_mode(_from)
	to, tstr := fm_to_mode(_to)
	if options.Config.Verbose > 1 {
		fmt.Printf("change <%s> => <%s>\n", fstr, tstr)
	}
	for _, v := range from {
		for _, m := range mranges {
			if uint16(m.boxid) == v {
				x.mchans[MODE_OFFSET+m.chanidx] = clrvalue(mranges, m)
			}
		}
	}

	for _, v := range to {
		for _, m := range mranges {
			if uint16(m.boxid) == v {
				x.mchans[MODE_OFFSET+m.chanidx] = setvalue(m)
			}
		}
	}
}

func (x *SitlGen) dump_chans(s string) {
	log.Printf("%-10.10s %+v\n", s, x.mchans)
}

func float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func float32tobytes(buf []byte, float float32) {
	bits := math.Float32bits(float)
	binary.LittleEndian.PutUint32(buf, bits)
}

func (x *SitlGen) sender(conn net.PacketConn, addr net.Addr, ch chan SimData) {
	buf := make([]byte, 1024)
	buf[0] = 'R'
	buf[1] = 'R'
	buf[2] = 'E'
	buf[3] = 'F'
	buf[4] = 0
	istart := 5
	sim := SimData{}
	for {
		select {
		case v := <-ch:
			sim = v
			istart = x.generate_buffer(buf, sim)
			_, err := conn.WriteTo(buf[:istart], addr)
			if err != nil {
				log.Printf("UDP write %v\n", err)
				return
			}
		case <-time.After(500 * time.Millisecond):
			istart = x.generate_buffer(buf, sim)
			_, err := conn.WriteTo(buf[:istart], addr)
			if err != nil {
				log.Printf("UDP write %v\n", err)
				return
			}
		}
	}
}

func (x *SitlGen) xplreader(conn net.PacketConn, achan chan net.Addr) {
	buf := make([]byte, 512)
	updatemap := true // Do not access a map in multiple threads
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err == nil {
			if n > 0 {
				ref := string(buf[0:4])
				if ref == "RREF" {
					freq := binary.LittleEndian.Uint32(buf[5:9])
					id := binary.LittleEndian.Uint32(buf[9:13])
					zb := bytes.Index(buf[13:], []byte("\000"))
					text := string(buf[13 : zb+13])
					if options.Config.Verbose > 2 {
						log.Printf("Read UDP %d %s %d %d %s\n", n, ref, freq, id, text)
					}
					parts := strings.Split(text, "/")
					item := parts[len(parts)-1]
					if item == "has_joystick" {
						updatemap = false
						achan <- addr
					}
					if updatemap {
						x.drefmap[item] = id
					}
				} else {
					//zb := bytes.Index(buf[9:], []byte("\000"))
					//text := string(buf[9 : zb+9])
					//                                              fval := float32frombytes(buf[5:9])
				}
			}
		} else {
			return
		}
	}
}

func (x *SitlGen) generate_buffer(buf []byte, sim SimData) int {
	istart := 5
	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["latitude"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Lat)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["longitude"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Lon)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["elevation"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Alt)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["groundspeed"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Speed)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["hpath"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Cog)
	istart += 4

	var inhg = to_hg(sim.Alt)
	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["barometer_current_inhg"])
	istart += 4
	float32tobytes(buf[istart:istart+4], inhg)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["g_axil"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Acc_x)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["g_side"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Acc_y)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["g_nrml"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Acc_z)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["P"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Gyro_x)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["Q"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Gyro_y)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["R"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Gyro_z)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["phi"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Roll)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["theta"])
	istart += 4
	float32tobytes(buf[istart:istart+4], -1.0*sim.Pitch)
	istart += 4

	binary.LittleEndian.PutUint32(buf[istart:istart+4], x.drefmap["psi"])
	istart += 4
	float32tobytes(buf[istart:istart+4], sim.Yaw)
	istart += 4
	return istart
}

func to_hg(alt float32) float32 {
	var k = 44330.0
	var p0 = 1013.25
	var p = p0 * (math.Pow(float64(1.0-float64(alt)/k), 5.255))
	return float32(0.0295299837 * p)
}

func proc_start(args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
}

func (x *SitlGen) arm_action(rxchan chan MSPChans, action bool) {
	if x.swchan != -1 {
		var act string
		if action {
			act = ""
			x.mchans[2] = 1997
			x.mchans[3] = 999
			x.mchans[x.swchan] = x.swval
		} else {
			act = "Dis"
			x.mchans[x.swchan] = 1002
			x.mchans[2] = 1500
			x.mchans[3] = 998
		}
		rxchan <- x.mchans
		if options.Config.Verbose > 0 {
			log.Printf("%sArming on chan %d at %d\n", act, x.swchan+1, x.mchans[x.swchan])
		}
	} else {
		log.Printf("No Arming switch, doomed\n")
	}
}

func log_mode_change(mranges []ModeRange, imodes []uint16, fname string) {
	var sb strings.Builder
	sb.WriteString("New mode <")
	sb.WriteString(fname)
	sb.WriteString("> ")
	sb.WriteString(fmt.Sprintf("%+v", imodes))
	sb.WriteString("\n")
	for _, r := range mranges {
		for _, k := range imodes {
			if uint16(r.boxid) == k {
				sb.WriteString(fmt.Sprintf(" %+v\n", r))
			}
		}
	}
	log.Printf(sb.String())
}

func (x *SitlGen) Run(rdrchan chan interface{}, meta types.FlightMeta) {
	log.SetPrefix("[fl2sitm] ")
	log.SetFlags(log.Ltime)
	args := read_cfg(options.Config.SitlEEprom)

	uaddr, err := net.ResolveUDPAddr("udp", options.Config.SitlListen)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", uaddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if proc, err := proc_start(args...); err == nil {
		defer func() {
			if options.Config.Verbose > 10 {
				log.Printf("DBG kill proc +%v\n", proc)
			}
			proc.Kill()
			proc.Wait()
		}()
	}

	var sim SimData

	// sim data to SITL
	simchan := make(chan SimData, 1)
	// socket addr, socket is open
	addrchan := make(chan net.Addr, 1)

	// RX data to simulator TX
	rxchan := make(chan MSPChans, 1)
	// RX status channel
	rxstat := make(chan byte, 1)
	rssich := make(chan byte, 1)

	// BBL data
	bbchan := make(chan SimData, 1)
	// BBL Command channel
	bbcmd := make(chan byte, 1)

	var m *MSPSerial = nil

	go x.xplreader(conn, addrchan)

	for j, _ := range x.mchans {
		x.mchans[j] = 0xffff
	}

	cnt := 0
	serial_ok := 0
	armed := false
	lastfm := uint16(types.FM_UNK)
	var mranges []ModeRange

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := keyboard.Close()
		if options.Config.Verbose > 10 {
			log.Printf("DBG reset k/bd +%v\n", err)
		}
	}()

	for done := false; done == false; {
		cnt += 1
		if options.Config.Verbose > 4 {
			log.Printf("Tick %d\n", cnt)
		}
		select {
		case addr := <-addrchan:
			if options.Config.Verbose > 1 {
				log.Printf("Got connection %+v\n", addr)
			}
			go x.sender(conn, addr, simchan)
			serial_ok = 1
			if options.Config.Verbose > 1 {
				log.Printf("Start BBL reader\n")
			}
			go file_reader(rdrchan, bbchan, bbcmd, float32(meta.Acc1G))
			sim = <-bbchan
			sim.Acc_x = 0.0
			sim.Acc_y = 0.0
			sim.Acc_z = 1.0
			simchan <- sim
		case <-time.After(100 * time.Millisecond):
			switch serial_ok {
			case 1:
				m, err = NewMSPSerial(options.Config.SitlPort)
				if err == nil {
					log.Printf("******** Opened RX **************\n")
					serial_ok = 2
				} else {
					log.Printf("Failed to open RX %v\n", err)
				}
			case 2:
				if options.Config.Verbose > 1 {
					log.Printf("Serial init\n")
				}
				go m.init(rxchan, rxstat, rssich)
				serial_ok = 3
				rssich <- sim.Rssi
			case 3:
				if options.Config.Verbose > 1 {
					log.Printf("Serial running %d\n", cnt)
				}
				rssich <- sim.Rssi
				serial_ok = 4
			default:
			}
		case sd := <-bbchan:
			simchan <- sd
			rssich <- sd.Rssi
			if options.Config.Verbose > 1 {
				log.Printf("SIM: %+v\n", sd)
			}
			if sd.Fmode == types.FM_UNK {
				done = true
				break
			}
			if sd.Fmode != lastfm {
				imodes, fname := fm_to_mode(sd.Fmode)
				if options.Config.Verbose > 1 {
					log_mode_change(mranges, imodes, fname)
				}
				x.change_mode(mranges, lastfm, sd.Fmode)
				if options.Config.Verbose > 1 {
					x.dump_chans("Mode")
				}
				rxchan <- x.mchans
				lastfm = sd.Fmode
			}

			if armed {
				x.mchans[0] = sd.RC_a
				x.mchans[1] = sd.RC_e
				x.mchans[2] = sd.RC_r
				x.mchans[3] = sd.RC_t
				rxchan <- x.mchans
			} else {
				x.arm_action(rxchan, true)
			}

		case rv := <-rxstat:
			if options.Config.Verbose > 1 {
				log.Printf("Status from rx %d\n", rv)
			}
			switch rv {
			case 1: /* ready to arm */
				if x.swchan == -1 {
					mranges = m.get_ranges()
					for _, r := range mranges {
						if r.boxid == PERM_ARM {
							x.swchan = 4 + int16(r.chanidx)
							x.swval = uint16(r.end+r.start)*25/2 + 900
							break
						}
					}
					log.Printf("** Ready to arm (Press 'A' to arm) **")
				}
			case 2: // disarm
				x.arm_action(rxchan, false)
				armed = false
			case 3: // armed
				log.Println("Armed")
				armed = true
				bbcmd <- 1 // awake reader
			case 0xff:
				done = true
				break
			default:
			}

		case ev := <-keysEvents:
			if ev.Err != nil {
				panic(ev.Err)
			}
			if ev.Key == 0 {
				switch ev.Rune {
				case 'A', 'a':
					x.arm_action(rxchan, true)
				case 'U':
					x.arm_action(rxchan, false)
				case 'Q', 'q':
					log.Println("Quit")
					done = true
				}
			} else if ev.Key == keyboard.KeyCtrlC {
				log.Println("Interrupt")
				done = true
			}

		}
	}

	if armed {
		log.Println("Disarming ...")
		x.arm_action(rxchan, false)
		armed = false
		for done := false; !done; {
			select {
			case <-rxstat:
				done = true
			case <-time.After(5000 * time.Millisecond):
				done = true
			}
		}
		log.Println("Cleanup ...")
		time.Sleep(2500 * time.Millisecond)
		log.Println("Done")
	}
}
