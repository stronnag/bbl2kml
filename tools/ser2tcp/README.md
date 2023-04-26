# ser2tcp

Here are some examples of a serial to TCP bridge, compatible with INAV Configurator's Ser2TCP / ser2TCP / Ser2TCP.exe.

Motivation.

* The INAV shipped Linux ser2TCP is **23MB**; yes 23MB for a simple bridge between two file descriptions. Which then doesn't work in the shipped version. The Linux compiled code here c. 23KB.
* There are no binary versions provided for other OS / Architectures such as MacOS/amd64 and Linux/ia32. Nor for architectures like Arm64 or  Riscv64.
* Curiosity. Solving the same problem in multiple languages.

The code here can be built for numerous OS and hardware, in particular the golang version.

## Options

### [./eg_c](eg_c)

C language. The smallest, fastest implementation. Fully functional on Linux, but cannot set non-standard baud rates on other OS. On Linux, requires `libudev` build environment (e.g. `libudev-dev` on Debian).

### [./eg_golang](eg_golang)

Golang. Excutables are c. 2MB and are static (no shared library requirements). Available for any hardware / OS for which there is a Go compiler.

### [./eg_rust](eg_rust)

Rust (2021 edition required). Excutables are c. 400KB. Available for any hardware / OS for which there is a Rust compiler.

## So which should I choose?

Your choice (if you care). For the record, the author uses the rust version.

## Usage

All language options offer the same command line options (compatible with the INAV version). 'Go' only offers the "long" options, 'C' and 'rust' offer both "long" and single character options.

```
$ ser2tcp  --help
ser2tcp [options]
Options:
    -h, --help          print this help menu
    -V, --version       print version and exit
    -v, --verbose       print I/O read sizes
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
