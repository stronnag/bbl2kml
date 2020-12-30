# bbl2kml

## Overview

Generate annotated KML/KMZ files from inav blackbox logs

```
$ ./bbl2kml --help
Usage of bbl2kml [options] file...
  -dms
    	Show positions as DMS (vice decimal degrees)
  -dump
    	Dump headers and exit
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -kml
    	Generate KML (vice KMZ)
  -rssi
    	Set RSSI view as default
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

### RSSI Track

* RSSI shading; range from red (100%) to yellow (0%), 10 step gradient

### Examples

![Example 1](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-1.png)

![Example 2](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-2.png)

![Example 3](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-3.png)

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

Binaries are provided for common operating systems in the [Release folder](https://github.com/stronnag/bbl2kml/releases) in
