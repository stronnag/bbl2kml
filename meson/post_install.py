#!/usr/bin/env python3
## python, so it (maybe) works on Windows
import platform
import os
import sys

exe=''
if platform.system() == "Windows":
    exe = ".exe"
    print("EXE from OS", file=sys.stderr)

if "EXE" in os.environ:
    exe = os.environ.get("EXE")
    print("EXE from ENV", file=sys.stderr)

inspath = os.path.join(os.environ['MESON_INSTALL_PREFIX'],'bin')
dst = os.path.join(inspath, 'fl2ltm'+exe)
src = os.path.join(inspath, 'fl2mqtt'+exe)

if os.path.exists(dst):
    os.remove(dst)
os.link(src, dst)

dst = os.path.join(inspath, 'bbsummary'+exe)
src = os.path.join(inspath, 'flightlog2kml'+exe)

if os.path.exists(dst):
    os.remove(dst)
os.link(src, dst)
