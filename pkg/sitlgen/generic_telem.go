package sitlgen

import (
	"log"
	"net"
	/**
		"encoding/binary"
		"fmt"
		"io"
		"os"
		"time"
	**/)

func GenericTelemReader(name string, tconn net.Conn) {
	inp := make([]byte, 256)
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	loc := tconn.LocalAddr().String()
	dst, err := net.ResolveUDPAddr("udp", loc)
	if err != nil {
		log.Fatal(err)
	}
	/*****************
	    var start time.Time
			w, err := os.Create(fmt.Sprintf("/tmp/%s-telem.log", name))
			if err == nil {
				defer w.Close()
				w.Write([]byte("v2\n"))
			} else {
				return
			}
		  **********************************/
	for {
		nb, err := tconn.Read(inp)
		if err == nil {
			conn.WriteTo(inp[0:nb], dst)
			/********************
						if start.IsZero() {
							start = time.Now()
						}
						diff := float64(time.Now().Sub(start)) / 1000000000.0
						var header = struct {
							offset float64
							size   uint16
							dirn   byte
						}{offset: diff, size: uint16(nb), dirn: 'i'}
						binary.Write(w, binary.LittleEndian, header)
						w.Write(inp[0:nb])
			      ************/
		} else {
			tconn.Close()
			return
		}
	}
}
