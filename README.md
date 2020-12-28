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
[blackbox_decode](https://github.com/iNavFlight/blackbox-tools)
