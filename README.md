# flightlog2kml

## Overview

Generate annotated KML/KMZ files from inav blackbox logs and OpenTX log files (inav S.Port telemetry).

* flightlog2kml - Generates KML/Z file(s) from Blackbox log(s) and OpenTX (OTX) logs
* mission2kml - Generate KML file from inav mission files (and other formats)
* fl2mqtt - Generates MQTT data to stimulate the on-line Ground Control Station [BulletGCSS](https://bulletgcss.fpvsampa.com/)

```
 flightlog2kml --help
Usage of flightlog2kml [options] file...
  -dms
    	Show positions as DD:MM:SS.s (vice decimal degrees) (default true)
  -dump
    	Dump log headers and exit
  -efficiency
    	Include efficiency layer in KML/Z (default true)
  -extrude
    	Extends track points to ground (default true)
  -gradient string
    	Specific colour gradient [red,rdgn,yor] (default "yor")
  -home-alt int
    	[OTX] home altitude
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -kml
    	Generate KML (vice default KMZ)
  -mission string
    	Optional mission file name
  -rssi
    	Set RSSI view as default
  -split-time int
    	[OTX] Time(s) determining log split, 0 disables (default 120)

flightlog2kml 0.8.4, commit: 0adaefb / 2021-01-09
```

Multiple logs (with multiple indices) may be given. A KML/Z will be generated for each file / index.

The output file is named from the base name of the source log file, appended with the index number and `.kml` or `.kmz` as appropriate. For example:

```
$ flightlog2kml LOG00044.TXT
Log      : LOG00044.TXT / 1
Flight   : "Model" on 2020-04-12T14:24:01.410+03:00
Firmware : INAV 2.4.0 (bcd4caef9) MATEKF722 of Feb 11 2020 22:48:59
Size     : 19.36 MB
Altitude : 292.8 m at 25:42
Speed    : 28.0 m/s at 13:54
Range    : 17322 m at 14:22
Current  : 30.6 A at 00:05
Distance : 48437 m
Duration : 43:44
Disarm   : Switch

results in the KMZ file "LOG00044.1.kmz"
```

Where `-mission <file>` is given, the given waypoint `<mission file>` will be included in the generated KML/Z; mission files may be one of the following formats as supported by [impload](https://github.com/stronnag/impload):

* MultiWii / XML mission files (MW-XML) ([mwp](https://github.com/stronnag/mwptools/), [inav-configurator](https://github.com/iNavFlight/inav-configurator), [ezgui](https://play.google.com/store/apps/details?id=com.ezio.multiwii&hl=en_GB), [mission planner for inav](https://play.google.com/store/apps/details?id=com.eziosoft.ezgui.inav&hl=en), drone-helper).
* [mwp JSON files](https://github.com/stronnag/mwptools/)
* [apmplanner2](https://ardupilot.org/planner2/) "QGC WPL 110" text files
* [qgroundcontrol](http://qgroundcontrol.com/) JSON plan files
* GPX and CSV (as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide))

If you use a format other than MW-XML or mwp JSON, it is recommended that you review any relevant format constraints as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide).

## Output

KML/Z file defining tracks which may be displayed Google Earth. Tracks can be animated with the time slider.

Both Flight Mode and RSSI tracks are generated; the default for display is Flight Mode, unless `-rssi` is specified (and RSSI data is available in the log). The log summary is displayed by double clicking on the "file name"` folder in Google Earth.

### Modes

`flightlog2kml` can generate three distinct colour-coded outputs:

* Flight mode: the default, colours as [below](#flight_mode_track).
* RSSI mode: RSSI percentage as a colour gradient, according to the current `--gradient` setting. Note that if no valid RSSI is found in the log, this mode will be suppressed.
* Efficiency mode: The efficiency (mAh/km) as a colour gradient,  according to the current `--gradient` setting. This is not enabled by default, and requires the `--efficiency` setting to be specified, either as a command line option or permanently in `$BBL2KML_OPTS`.

#### Flight Mode Track

* White : WP Mission
* Yellow : RTH
* Green : Pos Hold
* Lighter Green : Alt Hold
* Purple : Cruise
* Cyan : Piloted
* Lighter cyan : Launch
* Red : Failsafe
* Orange : Emergency Landing

### Colour Gradients

The RSSI and Efficiency modes are displayed using a colour gradient. Three gradients are available:
* `red` : The default, white representing the best (100%), red the worst (0%)
* `rdgn` : Red to green, green representing the best (100%), red the worst (0%)
* `yor` : Yellow/Orange/Red, yellow representing the best (100%), red the worst (0%)

If no option is given, `red` is assumed. Values are set by the `--gradient` command line option or  in `$BBL2KML_OPTS`.

### Examples

Note: These images are rather old, it looks much better now.

#### Flight Modes

![Example 1](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-1.png)

![Example 2](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-2.png)

![Example 3](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-3.png)

#### RSSI

![Example 4](https://github.com/stronnag/mwptools/wiki/images/inav-tracer-rssi.jpg)

## Using OpenTX logs

There are a few issues with OpenTX logs, the first of which needs OpenTX 2.3.11 (released 2021-01-08) to be resolved:
* CRSF logs in OpenTX 2.3.10 do not record the FM (Flight Mode) field. This makes it impossible to determine flight mode, or even if the craft is armed. Currently `flightlog2kml` tries to evince the armed state from other data.
* GPS Elevation. Unless you have a GPS attached to the TX, you don't get GPS altitude. This can be set by the `-home-alt H` value (in metres). Otherwise `flightlog2kml` will use an online elevation service.
* OpenTX creates a log per calendar day (IIRC), this means there may be multiple logs in the same file. Delimiting these individual logs is less than trivial, to some degree due to the prior CRSF issue which means arm / disarm is not reliably available. Currently, `flightlog2kml` assumes that a gap of more than 120 seconds indicates a new flight. The `-split-time` value allows a user-defined split time (seconds). Setting this to zero disables the log splitting function.


## `fl2mqtt`

The MQTT option (BulletGCSS) requires a MQTT broker URI, which may include a username/password and cafile if you require authentication and/or encryption.

```
$ fl2mqtt --help
Usage of fl2mqtt [options] file...
  -broker string
    	Mqtt URI (mqtt://[user[:pass]@]broker[:port]/topic[?cafile=file]
  -dump
    	Dump log headers and exit
  -home-alt int
    	[OTX] home altitude
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -mission string
    	Optional mission file name
  -rebase string
    	rebase all positions on lat,lon[,alt]
  -split-time int
    	[OTX] Time(s) determining log split, 0 disables (default 120)
```

The [BulletGCSS wiki](https://github.com/danarrib/BulletGCSS/wiki) describes how these values are chosen; in general:

* It is safe to use `broker.emqx.io` as the MQTT broker, this is default if no broker host is defined in the URI.
* You should use a unique topic for publishing your own data, this is slash separated string, for example `foo/bar/quux/demo`; the topic should include at least three elements.
* If you want to use a TLS (encrypted) connection to the broker, you may need to supply the broker's CA CRT (PEM) file. A reputable test broker will provide this via their web site.

Note that the scheme (**mqtt**:// in the `--help` text) is interpreted as:

* ws - Websocket (vice TCP socket), ensure the websocket port is also specificed
* wss - Encrypted websocket, ensure the TLS websocket port is also specificed. TLS validation is performed using the system CA files.
* mqtts - Secure (TLS) TCP connection. Ensure the TLS port is specified. TLS validation is performed using the system CA files.
* mqtt (or anyother scheme) - TCP connection. If `?cafile=file` is specified, then that is used for TLS validation (and the TLS port should be specified).


Example:

```
## the default broker is used ##
$ fl2mqtt -broker mqtt://broker.emqx.io/org/mwptools/mqtt/playotx openTXlog.csv
$ fl2mqtt -broker mqtt:///org/mwptools/mqtt/playbbl blackbox.TXT

## broker is test.mosquitto.org, over TLS,
## note the TLS port is also given (8883 in this case)
$ fl2mqtt -broker mqtt://test.mosquitto.org:8883/fl2mqtt/fl2mtqq/test?cafile=mosquitto.org.crt -mission simple_jump.mission BBL_102629.TXT
$ fl2mqtt -broker mqtts://test.mosquitto.org:8883/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
## Web sockets (plain text / TLS)
$ fl2mqtt -broker ws://test.mosquitto.org:8080/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
$ fl2mqtt -broker wss://test.mosquitto.org:8081/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
```

If a mission file is given, this will also be displayed by BulletGCSS, albeit incorrectly if there WP contains types other than `WAYPOINT` and `RTH`.


## `mission2kml`

A standalone mission file to KML/Z converter is also provided.

```
$ mission2kml --help
Usage of mission2kml [options] mission_file
  -dms
    	Show positions as DMS (vice decimal degrees)
  -home string
    	Use home location

The home location is given as decimal degrees latitude and
longitude and optional altitude. The values should be separated by a single
separator, one of "/:; ,". If space is used, then the values must be enclosed
in quotes.

In locales where comma is used as decimal "point", then comma should not be
used as a separator.

If a syntactically valid home position is given, without altitude, an online
elevation service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789/104
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236,24
```

A KML file is generated to stdout, which may be redirected to a file, e.g:

```
$ mission2kml -home 54.125229,-4.730443 barrule-h.mission > mtest.kml
```

## Setting default options

It is possible to define default options using the `BBL2KML_OPTS` (sic) environment variable.

```
BBL2KML_OPTS='-dms' flightlog2kml somelog.TXT
```

A permanent value can set in e.g. `.bashrc`, `.pam_environment` or Windows' equivalent.

```
export BBL2KML_OPTS='-dms -extrude'
or
export BBL2KML_OPTS='-rssi'
```

In the permanent usage case, options may be changed / inverted by command line using explicit values.

```
$ echo $BBL2KML_OPTS
-dms -extrude --gradient=yor --efficiency

$ flightlog2kml -extrude=false --dms=false randomBBL.TXT
```

The following options are recognised in `$BBL2KML_OPTS`; any other values (e.g. the obsolete `--elev` will cause the application to terminate. This is a feature.

* `--kml`
* `--rssi`
* `--extrude`
* `--gradient=red`
* `--decoder=blackbox_decode` The `blackbox_decode` application to use. This setting enables the use of experimental (or obsolete) decoders, mainly for testing and is thus only available via the environment.
* `--efficiency`

Note that the command interpreter allows `-flag` or `--flag` for any option.

## Limitations, Bugs, Bug Reporting

`flightlog2kml` aims to support as wide a range of inav firmware and log decoders as possible. During its development, inav has changed both the data logged and in some cases, the meaning of logged items; thus for versions of inav prior to 2.0, the reported flight mode might not be completely accurate. `flightlog2kml` is known to work with logs from 2015-10-30 (i.e. pre inav 1.0), and if you have a Blackbox log that is not decoded / visualised correctly, please raise a [Github issue](https://github.com/stronnag/bbl2kml/issues); this is a bug.

Due to the range of `inav` versions, `blackbox_decode` versions and supported operating systems, when reporting bugs, please include the following information in the Github issue:

* The version of `flightlog2kml` and `blackbox_decode`. Both applications have a `--help` option that should give the version numbers.
* The host operating system and version (e.g. "Debian Sid", "Windows 10", "MacOS 10.15").
* Provide the blackbox log that illustrates the problem. If you don't want to post the log into an essentially public forum (the Github issue), then please propose a private delivery channel.

## Building

Requires Go v1.13 or later.
Compiled with:

```
$ go build cmd/flightlog2kml/main.go
$ go build cmd/mission2kml/main.go
```

or more simply

```
make
```

**flightlog2kml** depends on [twpayne/go-kml](https://github.com/twpayne/go-kml), an outstanding Golang KML library.

`flightlog2kml` may be build for all OS for which a suitable Golang is available. It also requires inav's [blackbox_decode](https://github.com/iNavFlight/blackbox-tools); 0.4.5 (or future) is recommended; the minimum `blackbox_decode` version is 0.4.4. For Windows' users it is probably easiest to copy inav's `blackbox_decode.exe` into the same directory as `flightlog2kml.exe`.

Binaries are provided for common operating systems in the [Release folder](https://github.com/stronnag/bbl2kml/releases).
