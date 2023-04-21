package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.bug.st/serial"

	"strconv"
	"syscall"
)

var (
	device     string
	databits   int
	stopbits   string
	parity     string
	hostname   string
	port       int
	buffersize int
	baudrate   int
	verbose    bool
)

type SChan struct {
	ok   bool
	data []byte
}

func read_serial(s serial.Port, c0 chan SChan) {
	inp := make([]byte, 1024)
	for {
		n, err := s.Read(inp)
		if verbose {
			log.Printf("serial %d %v\n", n, err)
		}
		if err == nil {
			c0 <- SChan{ok: true, data: inp[0:n]}
		} else {
			c0 <- SChan{ok: false, data: []byte{}}
			return
		}
	}
}

func read_tcp(conn net.Conn, c0 chan SChan) {
	inp := make([]byte, 1024)
	for {
		n, err := conn.Read(inp)
		if verbose {
			log.Printf("tcp %d %v\n", n, err)
		}
		if err == nil {
			c0 <- SChan{ok: true, data: inp[0:n]}
		} else {
			c0 <- SChan{ok: false, data: []byte{}}
			return
		}
	}
}

func get_stop_bits() serial.StopBits {
	switch stopbits {
	case "Two":
		return serial.TwoStopBits
	case "OnePointFive":
		return serial.OnePointFiveStopBits
	}
	return serial.OneStopBit
}

func get_parity() serial.Parity {
	switch stopbits {
	case "Odd":
		return serial.OddParity
	case "Even":
		return serial.EvenParity
	case "Space":
		return serial.SpaceParity
	case "Mark":
		return serial.MarkParity
	}
	return serial.NoParity
}

func main() {
	vers := false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.BoolVar(&vers, "version", false, "Show version information")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose information")
	flag.StringVar(&device, "comport", "", "Serial device name")
	flag.IntVar(&baudrate, "baudrate", 115200, "Serial baud rate")
	flag.IntVar(&databits, "databits", 8, "Databits [5|6|7|8]")
	flag.StringVar(&stopbits, "stopbits", "One", "Stopbits [None|One|OnePointFive|Two]")
	flag.StringVar(&parity, "parity", "None", "Parity [Even|Mark|None|Odd|Space]")
	flag.StringVar(&hostname, "ip", "127.0.0.1", "Host name")
	flag.IntVar(&port, "tcpport", 5761, "IP port")
	flag.IntVar(&buffersize, "buffersize", 1024, "Buffersize")

	flag.Parse()
	if vers {
		fmt.Println("0.0.0")
		return
	}

	mode := &serial.Mode{
		BaudRate: baudrate,
		DataBits: databits,
		StopBits: get_stop_bits(),
		Parity:   get_parity(),
	}

	ser, err := serial.Open(device, mode)
	if err != nil {
		log.Fatal(err)
	}

	cc := make(chan os.Signal, 1)
	signal.Notify(cc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	remote := net.JoinHostPort(hostname, strconv.Itoa(port))
	host, err := net.ResolveTCPAddr("tcp", remote)

	if err != nil {
		println("ResolveTCP host failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, host)
	if err != nil {
		println("Connect failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	sc0 := make(chan SChan)
	tc0 := make(chan SChan)

	go read_serial(ser, sc0)
	go read_tcp(conn, tc0)

	for {
		select {
		case v := <-sc0:
			if v.ok {
				conn.Write(v.data)
			} else {
				return
			}
		case v := <-tc0:
			if v.ok {
				ser.Write(v.data)
			} else {
				return
			}
		case <-cc:
			log.Fatalln("Terminated")
			return
		}
	}
}
