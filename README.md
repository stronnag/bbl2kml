# bbl2kml

## Overview

Generate annotated KML/KMZ files from inav blackbox logs

```
$ ./bbl2kml --help
Usage of bbl2kml [options] file...
  -dump
    	Dump headers and exit
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms), default 100
  -kmz
    	Gnerate KMZ (vice KML)
```

Multiple logs (with multiple indices) may be given. A KML/Z will be generated for each file / index.

The output file is named from the base name of the Blackbox log file, appended with the index number and `.kml` or `.kmz` as appropriate. For example:

```
bbl2kml -kmz -interval 1000 /tmp/LOG00022.TXT
Log      : LOG00022.TXT / 1
Craft    :  on 2020-11-08T14:08:22.500+00:00
Fireware : INAV 2.3.0 (063ba5a) MATEKF722 of Jan 19 2020 20:20:56
Size     : 13.50 MB
Altitude : 553.3 m at 26:12
Speed    : 1631.1 m/s at 57:24
Range    : 22735 m at 27:58
Current  : 16.2 A at 00:10
Distance : 51899 m
Duration : 49:33
Disarm   : NONE

```
gives the KMZ file `LOG00022.1.kmz`

## Building

Compiled with

```
$ go build
```

or

```
make
```

bbl2kml may be build for all OS for which Golang is available. It also requires inav's
[blackbox_decode](https://github.com/iNavFlight/blackbox-tools). For Windows' users it is probablly easier to copy blackbox_decode.exe into the same directory as bbl2kml.exe.
