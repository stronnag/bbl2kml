package ltmgen

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
)

const (
	DevClass_NONE = iota
	DevClass_UDP
)

type DevDescription struct {
	klass  int
	name   string
	param  int
	name1  string
	param1 int
}

type MSPSerial struct {
	klass  int
	reader *bufio.Reader
	conn   net.Conn
}

func parse_device(device string, baud int) DevDescription {
	dd := DevDescription{name: "", klass: DevClass_NONE}
	if u, err := url.Parse(device); err == nil {
		dd.klass = DevClass_UDP
		dd.name = u.Hostname()
		p, _ := strconv.ParseInt(u.Port(), 10, 64)
		dd.param = int(p)
	}
	return dd
}

func check_device(device string, baud int) DevDescription {
	devdesc := parse_device(device, baud)
	if devdesc.name == "" {
		log.Fatalln("msgdev: No device available")
	} else {
		log.Printf("Using device [%v]\n", devdesc.name)
	}
	return devdesc
}

func (m *MSPSerial) Read(inp []byte) (int, error) {
	return m.reader.Read(inp)
}

func (m *MSPSerial) Write(payload []byte) (int, error) {
	return m.conn.Write(payload)
}

func (m *MSPSerial) Close() {
	m.conn.Close()
}

func (m *MSPSerial) Klass() int {
	return m.klass
}

func NewMSPSerial(device string, baud int) *MSPSerial {
	var laddr, raddr *net.UDPAddr
	var reader *bufio.Reader
	var conn net.Conn
	var err error
	dd := check_device(device, baud)
	if dd.name == "" {
		laddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", dd.name, dd.param))
	} else {
		raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", dd.name, dd.param))
	}
	if err == nil {
		conn, err = net.DialUDP("udp", laddr, raddr)
		if err == nil {
			reader = bufio.NewReader(conn)
		}
	}
	if err != nil {
		log.Fatalf("msgdev: %+v\n", err)
	}
	return &MSPSerial{klass: dd.klass, conn: conn, reader: reader}
}
