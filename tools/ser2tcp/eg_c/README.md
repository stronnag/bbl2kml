# Efficient ser2tcp

The `C` language example of a serial to TCP bridge, compatible with INAV Configurator's Ser2TCP / ser2TCP / Ser2TCP.exe and somewhat more efficient.

Motivation.

* The INAV shipped Linux ser2TCP is **23MB**; yes 23MB for a simple bridge between two file descriptors. Which then doesn't work in the shipped version. The Linux compiled code here c. 23KB.

Binaries by the downloaded from the [fl2sitl wiki](https://github.com/stronnag/bbl2kml/wiki/fl2sitl#images).
