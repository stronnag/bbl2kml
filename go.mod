module bbl2kml

go 1.20

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
	geo v1.0.0
	kmlgen v1.0.0
	log2mission v1.0.0
	ltmgen v1.0.0
	mission v1.0.0
	options v1.0.0
	otx v1.0.0
	sitlgen v1.0.0
	types v1.0.0
)

require (
	cli v1.0.0 // indirect
	github.com/bmizerany/perks v0.0.0-20230307044200-03f9df79da1e // indirect
	github.com/deet/simpleline v0.0.0-20140919022041-9d297ff784a2 // indirect
	github.com/eclipse/paho.mqtt.golang v1.5.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/mazznoer/csscolorparser v0.1.5 // indirect
	github.com/twpayne/go-kmz v0.0.0-20160614194227-165281381e72 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	inav v1.0.0 // indirect
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
