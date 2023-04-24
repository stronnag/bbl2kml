# ser2tcp

Drop-in replacement for `ser2TCP`, `Ser2TCP`, `Ser2TCP.exe` for use with the INAV SITL.

This native code replacement offers a number of benefits over the distributed file:

* Better performance
* c. 1/20 of the size (400KB v. 23MB).
* Written in Rust, so it must be better ...

## Usage

```
$ ser2tcp --help
Usage: ser2tcp [options]
Version: 0.0.0

Options:
    -h, --help          print this help menu
    -v, --version       print version and exit
    -c, --comport       serial device name (mandatory)
    -b, --baudrate <115200>
                        serial baud rate
    -d, --databits <8>  serial databits 5|6|7|8
    -s, --stopbits <One>
                        serial stopbits [None|One|Two]
    -p, --parity <None> serial parity [Even|None|Odd]
    -i, --ip <localhost>
                        Host name / Address
    -t, --tcpport <5761>
                        IP port
    -z, --buffersize    Buffersize (ignored)
```
## Installation

For use with the INAV Configurator, in the Configurator installation directory:

### Linux

```
cp ser2tcp resources/sitl/linux/Ser2TCP
cd resources/sitl/linux
rm -f ser2TCP && ln Ser2TCP ser2TCP
```

### Windows

```
copy ser2tcp.exe resources/sitl/windows/Ser2TCP.exe
```

### Building

See the `Makefile` for various `cargo` recipes.

## Licence

GPL v2 or later. (c) Jonathan Hudson 2023

## Notes

* The options are (mainly) bug for bug compatible with the tool shipped with the INAV Configurator
* If an invalid serial port is given, the a "reasonable" available port will be chosen.
