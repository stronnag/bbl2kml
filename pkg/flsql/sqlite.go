package flsql

import (
	"github.com/jmoiron/sqlx"
	"log"
	_ "modernc.org/sqlite"
	"os"
)

import (
	"options"
	"types"
)

const SCHEMA = `CREATE TABLE IF NOT EXISTS meta (id integer PRIMARY KEY, dtg timestamp with timestamp, duration real, mname text, firmware text, fwdate text, disarm int, flags int, motors int, servos int, sensors  int, acc1g int, features int, start int, end int);
CREATE TABLE IF NOT EXISTS logerrs (id integer PRIMARY KEY, errstr text);
CREATE TABLE IF NOT EXISTS misc (id integer, type text, content text);
CREATE TABLE IF NOT EXISTS logs(id integer, idx integer,
 stamp integer, lat double precision, lon double precision,
 alt  double precision, galt  double precision, spd  double precision,
 amps  double precision, volts double precision,
 hlat  double precision, hlon  double precision, vrange double precision,
 tdist double precision, effic double precision,
 energy double precision, whkm  double precision, whAcc double precision,
 qval  double precision, sval  double precision, aval  double precision,
 bval  double precision, fmtext Text, utc  timestamp, throttle integer,
 cse  integer, cog  integer, bearing integer, roll  integer, pitch integer, hdop  integer,
 ail  integer, ele  integer, rud  integer, thr integer,
 gyro_x integer, gyro_y integer, gyro_z integer, acc_x integer, acc_y integer, acc_z integer,
 fix  integer, numsat integer, fmode integer, rssi  integer, status integer, activewp integer,
 navmode integer, hwfail integer, windx integer, windy integer, windz integer);
create unique index if not exists logidx on logs (id,idx);`

const IMETA = `insert into meta (id, dtg, duration, mname,firmware,fwdate, disarm, flags, motors, servos, sensors, acc1g, features, start, end) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
const ISERR = `insert into logerrs (id, errstr) values ($1,$2)`
const ISMISC = `insert into misc (id, type, content) values ($1,$2,$3)`
const ILOG = `insert into logs (id, idx, stamp,lat,lon,alt,galt,spd,amps,volts,hlat,hlon,vrange,tdist,effic,energy,whkm,whAcc,qval,sval,aval,bval,fmtext,utc,throttle,cse,cog,bearing,roll,pitch,hdop,ail,ele,rud,thr,gyro_x,gyro_y,gyro_z,acc_x,acc_y,acc_z,fix,numsat,fmode,rssi,status,activewp,navmode,hwfail,windx,windy,windz) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45,$46,$47,$48,$49,$50,$51,$52)`

type DBL struct {
	db     *sqlx.DB
	tx     *sqlx.Tx
	stamp  uint64
	lstamp int64
}

func NewSQLliteDB(fn string) DBL {
	var d DBL
	var err error

	os.Remove(fn)

	d.db, err = sqlx.Open("sqlite", fn)
	if err != nil {
		log.Fatalf("db %+v\n", err)
	}

	if _, err = d.db.Exec(SCHEMA); err != nil {
		log.Fatalf("tables %+v\n", err)
	}
	return d
}

func (d *DBL) Reset() {
	d.stamp = 0
	d.lstamp = 0
}

func (d *DBL) WriteText(idx int, typ string, content string) {
	if err := d.tx.MustExec(ISMISC, idx, typ, content); err != nil {
		log.Fatalf("errors %+v\n", err)
	}
}

func (d *DBL) Writemeta(m types.FlightMeta) {
	if m.Craft == "" {
		m.Craft = "noname"
	}
	d.tx.MustExec(IMETA, m.Index, m.Date, m.Duration.Seconds(), m.Craft, m.Firmware,
		m.Fwdate, m.Disarm, m.Flags, m.Motors, m.Servos, m.Sensors, m.Acc1G, m.Features, m.Start, m.End)
}

func (d *DBL) Writeerr(idx int, errs string) {
	if err := d.tx.MustExec(ISERR, idx, errs); err != nil {
		log.Fatalf("errors %+v\n", err)
	}
}

func (d *DBL) Begin() {
	d.tx = d.db.MustBegin()
}

func (d *DBL) Commit() {
	if err := d.tx.Commit(); err != nil {
		log.Fatalf("commit %+v\n", err)
	}
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
	case types.FM_LAND:
		fms = 15
	case types.FM_CRUISE3D, types.FM_CRUISE2D:
		fms = 18
	case types.FM_PH:
		fms = 9
	case types.FM_AH:
		fms = 8
	case types.FM_EMERG:
		fms = 19
	default:
		fms = 0
	}
	return fms
}

func (d *DBL) Writelog(idx int, nx int, b types.LogItem) {
	stamp := int64(0)
	if nx == 0 {
		if options.MissionFile != "" {
			d.tx.MustExec(ISMISC, idx, "mission", options.MissionFile)
			options.MissionFile = ""
		}
		if options.GeoZone != "" {
			d.tx.MustExec(ISMISC, idx, "geozone", options.GeoZone)
			options.GeoZone = ""
		}
		d.stamp = b.Stamp
	}

	//if stamp < 0 {
	//log.Printf(":DBG: %+v %+v %+v %+v %+v \n", idx, nx, b.Stamp, d.stamp, stamp)
	//}
	stamp = int64(b.Stamp - d.stamp)
	if stamp < 0 {
		stamp = d.lstamp
	}

	d.lstamp = stamp
	ltmmode := ltm_flight_mode(b.Fmode)
	gnavmode := (uint(b.Navmode) | (uint(b.Navextra) << 8))

	d.tx.MustExec(ILOG, idx, nx, stamp,
		b.Lat,
		b.Lon,
		b.Alt,
		b.GAlt,
		b.Spd,
		b.Amps,
		b.Volts,
		b.Hlat,
		b.Hlon,
		b.Vrange,
		b.Tdist,
		b.Effic,
		b.Energy,
		b.Whkm,
		b.WhAcc,
		b.Qval,
		b.Sval,
		b.Aval,
		b.Bval,
		b.Fmtext,
		b.Utc,
		b.Throttle,
		b.Cse,
		b.Cog,
		b.Bearing,
		b.Roll,
		b.Pitch,
		b.Hdop,
		b.Ail,
		b.Ele,
		b.Rud,
		b.Thr,
		b.Gyro_x,
		b.Gyro_y,
		b.Gyro_z,
		b.Acc_x,
		b.Acc_y,
		b.Acc_z,
		b.Fix,
		b.Numsat,
		ltmmode,
		b.Rssi,
		b.Status,
		b.ActiveWP,
		gnavmode,
		b.HWfail,
		b.Wind[0],
		b.Wind[1],
		b.Wind[2])
}

func (d *DBL) Close() {
	d.db.Close()
}
