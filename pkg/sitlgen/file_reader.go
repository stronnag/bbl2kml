package sitlgen

import (
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	options "github.com/stronnag/bbl2kml/pkg/options"
	"log"
	"time"
)

func from_bbl(b types.LogItem, acc1g float32) SimData {
	sd := SimData{}
	sd.Lat = float32(b.Lat)
	sd.Lon = float32(b.Lon)
	sd.Alt = float32(b.Alt)
	sd.Galt = float32(b.GAlt)
	sd.Speed = float32(b.Spd)
	sd.Cog = float32(b.Cog)
	sd.Roll = float32(b.Roll)
	sd.Pitch = float32(b.Pitch)
	sd.Yaw = float32(b.Cse)
	sd.Gyro_x = float32(b.Gyro_x)
	sd.Gyro_y = float32(b.Gyro_y)
	sd.Gyro_z = float32(b.Gyro_z)
	sd.Acc_x = float32(b.Acc_x) / acc1g
	sd.Acc_y = float32(b.Acc_y) / acc1g
	sd.Acc_z = float32(b.Acc_z) / acc1g
	sd.RC_a = uint16(b.Ail)
	sd.RC_e = uint16(b.Ele)
	sd.RC_r = uint16(b.Rud)
	sd.RC_t = uint16(b.Thr)
	sd.Fmode = uint16(b.Fmode)
	sd.Rssi = b.Rssi
	return sd
}

func file_reader(rch chan interface{}, sdch chan SimData, cmdch chan byte, acc1g float32) {
	lt := uint64(0)
	var sd SimData
	done := false
	if options.Config.Verbose > 1 {
		log.Printf("Logreader with Acc1G = %.1f\n", acc1g)
	}
	for !done {
		select {
		case v := <-rch:
			switch v.(type) {
			case types.LogItem:
				b := v.(types.LogItem)
				ts := b.Stamp
				sd = from_bbl(b, acc1g)
				sdch <- sd
				if lt != 0 {
					tdiff := time.Duration(b.Stamp-lt) * time.Microsecond
					if options.Config.Verbose > 1 {
						log.Printf("Reader sleeps %v\n", tdiff)
					}
					time.Sleep(tdiff)
				} else {
					if options.Config.Verbose > 11 {
						log.Println("Reader waits on cmd")
					}
					<-cmdch
					if options.Config.Verbose > 11 {
						log.Println("Reader continues on cmd")
					}
					sdch <- sd
				}
				lt = ts
			case types.HomeRec:
				done = false
			case types.MapRec:
				done = true
			}
		case <-time.After(1 * time.Millisecond):
			//
		}
	}
	if options.Config.Verbose > 1 {
		log.Printf("Reader EOF\n")
	}
	sd.Fmode = types.FM_UNK
	sdch <- sd
}
