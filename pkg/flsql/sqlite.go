package flsql

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"os"
)

import (
	"types"
)

const SCHEMA = `CREATE TABLE IF NOT EXISTS meta (id integer NOT NULL PRIMARY KEY, dtg timestamp with time zone, duration integer);
CREATE TABLE IF NOT EXISTS logerrs (id integer NOT NULL PRIMARY KEY, errstr text);
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
 navMode integer, hwfail integer, windx integer, windy integer, windz integer)`

const IMETA = `insert into meta (id, dtg, duration) values ($1,$2,$3)`
const ISERR = `insert into logerrs (id, errstr) values ($1,$2)`
const ILOG = `insert into logs (id,idx, stamp,lat,lon,alt,galt,spd,amps,volts,hlat,hlon,vrange,tdist,effic,energy,whkm,whAcc,qval,sval,aval,bval,fmtext,utc,throttle,cse,cog,bearing,roll,pitch,hdop,ail,ele,rud,thr,gyro_x,gyro_y,gyro_z,acc_x,acc_y,acc_z,fix,numsat,fmode,rssi,status,activewp,navmode,hwfail,windx,windy,windz) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45,$46,$47,$48,$49,$50,$51,$52)`

type DBL struct {
	db    *sql.DB
	count int
	stamp uint64
}

func NewSQLliteDB(fn string) DBL {
	var d DBL
	var err error

	os.Remove(fn)

	d.db, err = sql.Open("sqlite", fn)
	if err != nil {
		log.Fatalf("db %+v\n", err)
	}

	if _, err = d.db.Exec(SCHEMA); err != nil {
		log.Fatalf("tables %+v\n", err)
	}
	return d
}

func (d *DBL) Reset() {
	d.count = 0
	d.stamp = 0
}

func (d *DBL) Writemeta(m types.FlightMeta) {
	if _, err := d.db.Exec(IMETA, m.Index, m.Date, m.Duration.Seconds()); err != nil {
		log.Fatalf("meta %+v\n", err)
	}
}

func (d *DBL) Writeerr(idx int, errs string) {
	if _, err := d.db.Exec(ISERR, idx, errs); err != nil {
		log.Fatalf("errors %+v\n", err)
	}
}

func (d *DBL) Begin() {
	if _, err := d.db.Exec(`BEGIN TRANSACTION`); err != nil {
		log.Fatalf("begin %+v\n", err)
	}
}

func (d *DBL) Commit() {
	if _, err := d.db.Exec(`COMMIT`); err != nil {
		log.Fatalf("commit %+v\n", err)
	}
}

func (d *DBL) Writelog(idx int, b types.LogItem) {
	var stamp uint64
	if d.count == 0 {
		d.stamp = b.Stamp
	}
	stamp = b.Stamp - d.stamp

	_, err := d.db.Exec(ILOG, idx, d.count,
		stamp,
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
		b.Fmode,
		b.Rssi,
		b.Status,
		b.ActiveWP,
		b.NavMode,
		b.HWfail,
		b.Wind[0],
		b.Wind[1],
		b.Wind[2])
	if err != nil {
		log.Fatalf("log %+v\n", err)
	}
	d.count++
}

func (d *DBL) Close() {
	d.db.Close()
}
