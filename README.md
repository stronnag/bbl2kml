# bbl2kml

## Overview

Generate annotated KML/KMZ files from inav blackbox logs

```
$ bbl2kml --help
Usage of bb2kml [options] file...
  -dms
    	Show positions as DMS (vice decimal degrees)
  -dump
    	Dump headers and exit
  -elev
    	Use online elevation service to adjust mission evelations
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -kml
    	Generate KML (vice KMZ)
  -mission string
    	Mission file name
  -rssi
    	Set RSSI view as default

bbl2kml 0.2.0, commit: 816eea0 / 2021-01-02
```

Multiple logs (with multiple indices) may be given. A KML/Z will be generated for each file / index.

The output file is named from the base name of the Blackbox log file, appended with the index number and `.kml` or `.kmz` as appropriate. For example:

```
bbl2kml /tmp/LOG00022.TXT
Log      : LOG00022.TXT / 1
Craft    :  on 2020-11-08T14:08:22.500+00:00
Firmware : INAV 2.3.0 (063ba5a) MATEKF722 of Jan 19 2020 20:20:56
Size     : 13.50 MB
Altitude : 553.3 m at 26:12
Speed    : 23.7 m/s at 57:24
Range    : 22735 m at 27:58
Current  : 16.2 A at 00:10
Distance : 51899 m
Duration : 49:33
Disarm   : NONE

results in the KMZ file "LOG00022.1.kmz"
```

Where `-mission <file>` is given, the given waypoint `<mission file>` will be included in the generated KML/Z; mission files may be one of the following formats as supported by [impload](https://github.com/stronnag/impload):

* MultiWii / XML mission files (MW-XML) ([mwp](https://github.com/stronnag/mwptools/), [inav-configurator](https://github.com/iNavFlight/inav-configurator), [ezgui](https://play.google.com/store/apps/details?id=com.ezio.multiwii&hl=en_GB), [mission planner for inav](https://play.google.com/store/apps/details?id=com.eziosoft.ezgui.inav&hl=en), drone-helper).
* [mwp JSON files](https://github.com/stronnag/mwptools/)
* [apmplanner2](https://ardupilot.org/planner2/) "QGC WPL 110" text files
* [qgroundcontrol](http://qgroundcontrol.com/) JSON plan files
* GPX and CSV (as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide))

If you use a format other than MW-XML or mwp JSON, it is recommended that you review any relevant format constraints as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide).

## Output

KML/Z file defining tracks whch may be displayed Google Earth. Tracks can be animated with the time slider.

Both Flight Mode and RSSI tracks are generated; the default for display is Flight Mode, unless `-rssi` is specified (and RSSI data is available in the log). The log summary is displayed by double clicking on the `inav flight` folder in Google Earth.

### Flight Mode Track

* White : WP Mission
* Yellow : RTH
* Green : Pos Hold
* Lighter Green : Alt Hold
* Purple : Cruise
* Cyan : Piloted
* Lighter cyan : Launch
* Red : Failsafe
* Orange : Emergency Landing

### RSSI Track

* RSSI shading; range from red (100%) to yellow (0%), 10 step gradient

### Examples

#### Flight Modes

![Example 1](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-1.png)

![Example 2](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-2.png)

![Example 3](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-3.png)

#### RSSI

![Example 4](https://github.com/stronnag/mwptools/wiki/images/inav-tracer-rssi.jpg)

## Building

Compiled with:

```
$ go build
```

or

```
make
```

**bbl2kml** depends on [twpayne/go-kml](https://github.com/twpayne/go-kml), an outstanding Golang KML library.

bbl2kml may be build for all OS for which Golang is available. It also requires inav's
[blackbox_decode](https://github.com/iNavFlight/blackbox-tools); 0.4.5 (including RCs) or later is recommended; the minimum `blackbox_decode` version is 0.4.4. For Windows' users it is probably easiest to copy inav's `blackbox_decode.exe` into the same directory as `bbl2kml.exe`.

Binaries are provided for common operating systems in the [Release folder](https://github.com/stronnag/bbl2kml/releases).

## `mission2kml`

A standalone mission file to KML/Z converter is available in the repository; it is not built by default, but may be built from the `Makefile`, `make all`

```
$ mission2kml --help
Usage of missionkml [options] mission_file
  -dms
    	Show positions as DMS (vice decimal degrees)
  -home string
    	Use home location

The home location is given as decimal degrees latitude and
longitude. The values should be separated by a single separator, one
of "/:; ,". If space is used, then the values must be enclosed in
quotes. In locales where comma is used as decimal "point", then it
should not be used as a separator.

If a syntactically valid home postion is given, an online elevation
service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236

```

## Limitations, Bugs, Bug Reporting

`bbl2kml` aims to support as wide a range of inav firmware and log decoders as possible. During its development, inav has changed both the data logged and in some cases, the meaning of logged items; thus for versions of inav prior to 2.0, the reported flight mode might not be completely accurate. `bbl2kml` is known to work with logs from 2015-10-30 (i.e. pre inav 1.0), and if you have a Blackbox log that is not decoded / visualisated correctly, please raise a [Github issue](https://github.com/stronnag/bbl2kml/issues); this is a bug.

Due to the range of `inav` versions, `blackbox_decode` versions and supported operating systems, when reporting bugs, please include the following information in the Github issue:

* The version of `bbl2kml` and `blackbox_decode`. Both applications have a `--help` option that should give the version numbers.
* The host operating system and version (e.g. "Debian Sid", "Windows 10", "MacOS 10.15").
* Provide the blackbox log that illustrates the problem. If you don't want to post the log into an essentially public forum (the Github issue), then please propose a private delivery channel.
