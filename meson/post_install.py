#!/usr/bin/env python3
## python, so it (maybe) works on Windows
import os

inspath = os.path.join(os.environ['MESON_INSTALL_PREFIX'],'bin')
ltm = os.path.join(inspath, 'fl2ltm')
mqtt = os.path.join(inspath, 'fl2mqtt')

if os.path.exists(ltm):
    os.remove(ltm)

os.link(mqtt, ltm)
