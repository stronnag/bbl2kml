# Efficient ser2tcp

The `C` language example of a serial to TCP bridge, compatible with INAV Configurator's Ser2TCP / ser2TCP / Ser2TCP.exe and somewhat more efficient.

## Motivation.

* The INAV shipped Linux ser2TCP is **23MB**; yes 23MB for a simple bridge between two file descriptors. Which then didn't work in the originally shipped version. The Linux compiled code here is c. 23KB (c.170KB static).

Binaries may be downloaded from the [fl2sitl wiki](https://github.com/stronnag/bbl2kml/wiki/fl2sitl#images) or built from source.

``` bash
$ make
# or FreeBSD
$ gmake
```

## Linux and non-standard baud rates

The baud rates available in Linux are the traditional POSIX rates, plus some additional multipliers in drivers. Some devices may support additional "non-standard" rates, however this is driver dependent.

* `ser2tcp` will warn when a baud rate cannot be supported precisely and the driver imposes an alternate rate.
* The optional `sertest` program will test the baud rate capability of the serial device.

### `sertest`

```
make sertest
```

If no arguments are given, `/dev/ttyUSB0` is tested at 57600, 100000, 115200, 200000, 230400, 400000, 420000, 460800 baud. For example:

```
$ ./sertest
OK /dev/ttyUSB0 : 57600
OK /dev/ttyUSB0 : 100000
OK /dev/ttyUSB0 : 115200
OK /dev/ttyUSB0 : 200000
OK /dev/ttyUSB0 : 230400
OK /dev/ttyUSB0 : 400000
Warning: device speed 421052 differs from requested 420000
OK /dev/ttyUSB0 : 420000
OK /dev/ttyUSB0 : 460800
```

Otherwise, a device name and baud rate(s) may be specified (e.g. on FreeBSD):

```
$ ./sertest /dev/cuaU0
OK /dev/cuaU0 : 57600
OK /dev/cuaU0 : 100000
OK /dev/cuaU0 : 115200
OK /dev/cuaU0 : 200000
OK /dev/cuaU0 : 230400
OK /dev/cuaU0 : 400000
OK /dev/cuaU0 : 420000
OK /dev/cuaU0 : 460800
$ ./sertest /dev/cuaU0 9600 19200 300 75
OK /dev/cuaU0 : 9600
OK /dev/cuaU0 : 19200
OK /dev/cuaU0 : 300
OK /dev/cuaU0 : 75
```

The FreeBSD driver supports arbitrary baud rates better than Linux does (as do the MacOS and Windows drivers).

Note that the Windows build requires "cygwin" and may use either MSDOS or pseudo-Unix style device names, e.g. `sertest COM5` and `sertest /dev/ttyS4` do the same thing.

```
$ ./sertest COM5
OK /dev/ttyS4 : 57600
OK /dev/ttyS4 : 100000
OK /dev/ttyS4 : 115200
OK /dev/ttyS4 : 200000
OK /dev/ttyS4 : 230400
OK /dev/ttyS4 : 400000
OK /dev/ttyS4 : 420000
OK /dev/ttyS4 : 460800
```
