package sqlreader

import (
	"github.com/jmoiron/sqlx"
	"log"
	_ "modernc.org/sqlite"
	"path/filepath"
	"time"
)

import (
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
	if err != nil {
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

func (o *SQLREAD) metas(logfile string) ([]types.FlightMeta, error) {
	var metas []types.FlightMeta
	bp := filepath.Base(logfile)

	rows, err := o.db.Query("SELECT * FROM meta order by id")
	if err == nil {
		for rows.Next() {
			m := types.FlightMeta{Logname: bp, Start: 1}
			var dt string
			var tm string
			err := rows.Scan(&m.Index, &tm, &dt, &m.Craft, &m.Firmware)
			if err != nil {
				log.Printf("META SQL: %+v\n", err)
				return metas, err
			}
			dt = dt + "s"
			m.Date, err = time.Parse("2006-01-02 15:04:05.999 -0700 MST", tm)
			m.Duration, err = time.ParseDuration(dt)
			m.Flags = types.Has_Start | types.Is_Valid | types.Has_Craft
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
	)

	rows, err := lg.db.Query("SELECT * FROM logs where id=$1 order by idx", m.Index)
	if err != nil {
		log.Fatalf("METASQL for %d +%v\n", m.Index, err)
	}
	for rows.Next() {
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
			&b.NavMode,
			&b.HWfail,
			&b.Wind[0],
			&b.Wind[1],
			&b.Wind[2])

		if err != nil {
			log.Printf("META SQL: %+v\n", err)
			break
		}

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
