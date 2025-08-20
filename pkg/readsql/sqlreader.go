package sqlreader

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import (
	"options"
	"types"
)

type SQLREAD struct {
	name string
	meta []types.FlightMeta
	db   *sqlx.DB
}

func NewSQLReader(fn string) SQLREAD {
	var l SQLREAD
	l.name = fn
	l.meta = nil
	l.db, _ = sqlx.Open("sqlite", fn)
	return l
}

func (o *SQLREAD) LogType() byte {
	return types.LOGSQL
}

func (o *SQLREAD) GetDurations() {
}

func (o *SQLREAD) Dump() {
}

func (o *SQLREAD) GetMetas() ([]types.FlightMeta, error) {
	m, err := types.ReadMetaCache(o.name)
	if err != nil || options.Config.Nocache {
		m, err = o.metas(o.name)
		types.WriteMetaCache(o.name, m)
	}
	o.meta = m
	return m, err
}

func fm_ltm(ltm uint8) uint8 {
	var fm uint8
	switch ltm {
	case 0:
		fm = types.FM_MANUAL
	case 1:
		fm = types.FM_ACRO
	case 2:
		fm = types.FM_ANGLE
	case 3:
		fm = types.FM_HORIZON
	case 8:
		fm = types.FM_AH
	case 9:
		fm = types.FM_PH
	case 10:
		fm = types.FM_WP
	case 13:
		fm = types.FM_RTH
	case 15:
		fm = types.FM_LAND
	case 18:
		fm = types.FM_CRUISE3D
	case 19:
		fm = types.FM_EMERG
	case 20:
		fm = types.FM_LAUNCH
	default:
		fm = types.FM_ACRO
	}
	return fm
}

func parse_time(tm string) (time.Time, error) {
	dt, err := time.Parse("2006-01-02 15:04:05.999 -0700 MST", tm)
	if err != nil {
		parts := strings.Split(tm, " ")
		if len(parts) == 4 {
			tm = strings.Join(parts[0:3], " ")
		} else if len(parts) == 2 {
			tm = tm + " +0000"
		}
		dt, err = time.Parse("2006-01-02 15:04:05.999 -0700", tm)
	}
	return dt, err
}

func (o *SQLREAD) metas(logfile string) ([]types.FlightMeta, error) {
	var metas []types.FlightMeta
	bp := filepath.Base(logfile)

	rows, err := o.db.Queryx("SELECT * FROM meta order by id")
	if err == nil {
		for rows.Next() {
			m := types.FlightMeta{Logname: bp, Start: 1}
			var dt string
			var tm string
			res := make(map[string]interface{})
			err = rows.MapScan(res)
			id := res["id"].(int64)
			m.Index = int(id)
			tm = res["dtg"].(string)
			dtf := res["duration"].(float64)
			m.Craft = res["mname"].(string)
			m.Firmware = res["firmware"].(string)
			m.Fwdate = res["fwdate"].(string)
			m.Disarm = types.Reason(res["disarm"].(int64))
			m.Flags = uint8(res["flags"].(int64))
			m.Motors = uint8(res["motors"].(int64))
			m.Servos = uint8(res["servos"].(int64))
			m.Sensors = uint16(res["sensors"].(int64))
			m.Acc1G = uint16(res["acc1g"].(int64))
			m.Features = uint32(res["features"].(int64))
			if tmp, ok := res["start"]; ok {
				m.Start = int(tmp.(int64))
			} else {
				m.Start = 1
			}
			if tmp, ok := res["end"]; ok {
				m.End = int(tmp.(int64))
			} else {
				m.End = 999999
			}
			if err != nil {
				log.Printf("META SQL: %+v\n", err)
				return metas, err
			}
			dt = fmt.Sprintf("%fs", dtf)
			m.Date, err = parse_time(tm)
			m.Duration, err = time.ParseDuration(dt)
			m.Flags |= types.Has_Start | types.Is_Valid | types.Has_Craft
			metas = append(metas, m)
		}
	}
	return metas, err
}

func (lg *SQLREAD) Reader(m types.FlightMeta, ch chan interface{}) (types.LogSegment, bool) {
	stats := types.LogStats{}
	ls := types.LogSegment{}
	rec := types.LogRec{}
	homes := types.HomeRec{}
	var (
		mid     int
		midx    int
		ltmmode int
		mstr    string
	)

	dname, err := os.MkdirTemp("", ".fl2kml.")
	fname := filepath.Join(dname, ".tmpmission.mission")
	f, err := os.Create(fname)
	if err == nil {
		rows, err := lg.db.Query("SELECT content FROM misc WHERE id = $1 and type = $2;", m.Index, "mission")
		if err == nil {
			for rows.Next() {
				err := rows.Scan(&mstr)
				if err == nil {
					f.WriteString(mstr)
				}
			}
			options.Config.Mission = fname
		}
		f.Close()
	}

	fname = filepath.Join(dname, ".tmpcli.txt")
	f, err = os.Create(fname)
	if err == nil {
		rows, err := lg.db.Query("SELECT content FROM misc WHERE id = $1 and type = $2;", m.Index, "climisc")
		if err == nil {
			for rows.Next() {
				err := rows.Scan(&mstr)
				if err == nil {
					f.WriteString(mstr)
				}
			}
		}
		options.Config.Cli = fname
		f.Close()
	}

	types.TDir = dname

	rows, err := lg.db.Query("SELECT * FROM logs where id=$1 order by idx", m.Index)
	if err != nil {
		log.Fatalf("METASQL for %d +%v\n", m.Index, err)
	}
	for rows.Next() {
		var navmode uint

		b := types.LogItem{}
		err := rows.Scan(&mid, &midx,
			&b.Stamp,
			&b.Lat,
			&b.Lon,
			&b.Alt,
			&b.GAlt,
			&b.Spd,
			&b.Amps,
			&b.Volts,
			&b.Hlat,
			&b.Hlon,
			&b.Vrange,
			&b.Tdist,
			&b.Effic,
			&b.Energy,
			&b.Whkm,
			&b.WhAcc,
			&b.Qval,
			&b.Sval,
			&b.Aval,
			&b.Bval,
			&b.Fmtext,
			&b.Utc,
			&b.Throttle,
			&b.Cse,
			&b.Cog,
			&b.Bearing,
			&b.Roll,
			&b.Pitch,
			&b.Hdop,
			&b.Ail,
			&b.Ele,
			&b.Rud,
			&b.Thr,
			&b.Gyro_x,
			&b.Gyro_y,
			&b.Gyro_z,
			&b.Acc_x,
			&b.Acc_y,
			&b.Acc_z,
			&b.Fix,
			&b.Numsat,
			&ltmmode,
			&b.Rssi,
			&b.Status,
			&b.ActiveWP,
			&navmode,
			&b.HWfail,
			&b.Wind[0],
			&b.Wind[1],
			&b.Wind[2])

		if err != nil {
			log.Printf("META SQL: %+v\n", err)
			break
		}

		b.Navmode = byte(navmode & 0xff)
		b.Fmode = fm_ltm(uint8(ltmmode))

		if b.Vrange > stats.Max_range {
			stats.Max_range = b.Vrange
			stats.Max_range_time = b.Stamp
		}

		if b.Alt > stats.Max_alt {
			stats.Max_alt = b.Alt
			stats.Max_alt_time = b.Stamp
			rec.Cap |= types.CAP_ALTITUDE
		}

		if b.Spd > 0 && b.Spd < 400 {
			if b.Spd > stats.Max_speed {
				stats.Max_speed = b.Spd
				stats.Max_speed_time = b.Stamp
				rec.Cap |= types.CAP_SPEED
			}
		}

		if b.Amps > stats.Max_current {
			stats.Max_current = b.Amps
			stats.Max_current_time = b.Stamp
			rec.Cap |= (types.CAP_VOLTS | types.CAP_AMPS)
		}

		if homes.Flags == 0 {
			if (b.Status & 1) != 0 {
				homes.Flags |= types.HOME_ARM | types.HOME_ALT
				homes.HomeLat = b.Hlat
				homes.HomeLon = b.Hlon
				homes.HomeAlt = b.GAlt
			}
		}

		if b.Rssi > 0 {
			rec.Cap |= types.CAP_RSSI_VALID
		}

		if b.Fmode == types.FM_WP {
			rec.Cap |= types.CAP_WPNO
		}

		stats.Distance = b.Tdist / 1852.0
		if ch != nil {
			ch <- b
		} else {
			rec.Items = append(rec.Items, b)
		}
	}

	stats.Max_range /= 1852.0
	srec := stats.Summary(uint64(m.Duration.Microseconds()))

	if ch != nil {
		ch <- srec
		return ls, true
	} else {
		ok := homes.Flags != 0 && len(rec.Items) > 0
		if ok {
			ls.L = rec
			ls.H = homes
			ls.M = srec
		}
		return ls, ok
	}
}
