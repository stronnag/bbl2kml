#!/bin/sh

rm -f $MESON_INSTALL_PREFIX/bin/fl2ltm
ln $MESON_INSTALL_PREFIX/bin/fl2mqtt $MESON_INSTALL_PREFIX/bin/fl2ltm
