#!/usr/bin/env python3
## python, so it (maybe) works on Windows
import os

inspath = os.path.join(os.environ['MESON_INSTALL_PREFIX'],'bin')
dst = os.path.join(inspath, 'fl2ltm')
src = os.path.join(inspath, 'fl2mqtt')

if os.path.exists(dst):
    os.remove(dst)
os.link(src, dst)

dst = os.path.join(inspath, 'bbsummary')
src = os.path.join(inspath, 'flightlog2kml')

if os.path.exists(dst):
    os.remove(dst)
os.link(src, dst)
