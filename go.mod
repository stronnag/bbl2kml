module bbl2kml

go 1.24.0

toolchain go1.24.5

require (
	github.com/mazznoer/colorgrad v0.10.0
	github.com/twpayne/go-kml v1.5.2
	github.com/yookoala/realpath v1.0.0
)

require (
	aplog v1.0.0
	bbl v1.0.0
	bltlog v1.0.0
	bltmqtt v1.0.0
	flsql v1.0.0
	geo v1.0.0
	kmlgen v1.0.0
	log2mission v1.0.0
	ltmgen v1.0.0
	mission v1.0.0
	mwpjson v1.0.0
	options v1.0.0
	otx v1.0.0
	sitlgen v1.0.0
	sqlreader v1.0.0
	types v1.0.0
)

require (
	cli v1.0.0 // indirect
	github.com/bmizerany/perks v0.0.0-20230307044200-03f9df79da1e // indirect
	github.com/deet/simpleline v0.0.0-20140919022041-9d297ff784a2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eclipse/paho.mqtt.golang v1.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/mazznoer/csscolorparser v0.1.6 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/twpayne/go-kmz v0.0.0-20160614194227-165281381e72 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	inav v1.0.0 // indirect
	modernc.org/libc v1.66.7 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.38.2 // indirect
	styles v1.0.0 // indirect
)

replace bbl v1.0.0 => ./pkg/bbl

replace bltlog v1.0.0 => ./pkg/bltreader

replace bltmqtt v1.0.0 => ./pkg/bltmqtt

replace geo v1.0.0 => ./pkg/geo

replace log2mission v1.0.0 => ./pkg/log2mission

replace options v1.0.0 => ./pkg/options

replace otx v1.0.0 => ./pkg/otx

replace inav v1.0.0 => ./pkg/inav

replace mission v1.0.0 => ./pkg/mission

replace types v1.0.0 => ./pkg/types

replace aplog v1.0.0 => ./pkg/aplog

replace ltmgen v1.0.0 => ./pkg/ltmgen

replace kmlgen v1.0.0 => ./pkg/kmlgen

replace sitlgen v1.0.0 => ./pkg/sitlgen

replace styles v1.0.0 => ./pkg/styles

replace cli v1.0.0 => ./pkg/cli

replace flsql v1.0.0 => ./pkg/flsql

replace mwpjson v1.0.0 => ./pkg/mwpjson/

replace sqlreader v1.0.0 => ./pkg/readsql/
