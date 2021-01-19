LDFLAGS ?= -s -w
prefix ?= /usr

_CAPP=flightlog2kml
_GAPP=fl2kmlgtk
_MAPP=mission2kml
_FAPP=fl2kmlfyne
_QAPP=fl2mqtt

ifndef DESTDIR
 CAPP=flightlog2kml
 GAPP=fl2kmlgtk
 MAPP=mission2kml
 FAPP=fl2kmlfyne
 QAPP=fl2mqtt
else
 CAPP=$(DESTDIR)/flightlog2kml
 MAPP=$(DESTDIR)/mission2kml
 GAPP=$(DESTDIR)/fl2kmlgtk
 FAPP=$(DESTDIR)/fl2kmlfyne
 QAPP=$(DESTDIR)/fl2mqtt
endif

ifeq ($(GOOS),windows)
 EXT=.exe
 LDEXTRA=-H=windowsgui
else
 EXT=
endif

all: $(_CAPP) $(_MAPP) $(_QAPP)

gui: $(_FAPP) $(_GAPP)

PKGCOMMON = $(wildcard pkg/api/*/*.go) $(wildcard pkg/mission/*.go) $(wildcard pkg/geo/*.go)
PKGOPT = $(wildcard pkg/options/*.go)
PKGBBL = $(wildcard pkg/bbl/*.go)
PKGOTX = $(wildcard pkg/otx/*.go)
PKGINAV = $(wildcard pkg/inav/*.go)
PKGKML = $(wildcard pkg/kmlgen/*.go)
PKGMQTT = $(wildcard pkg/bltmqtt/*.go)

CSRCS = $(wildcard cmd/flightlog2kml/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)
QSRCS = $(wildcard cmd/fl2mqtt/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGMQTT)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(PKGCOMMON)
GSRCS = cmd/fl2kmlgtk/main.go cmd/fl2kmlgtk/logkml.ui $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)
FSRCS = cmd/fl2kmlfyne/main.go $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)

$(_CAPP): $(CSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(CAPP)$(EXT) cmd/flightlog2kml/main.go

$(_MAPP): $(MSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

$(_GAPP): $(GSRCS) cmd/fl2kmlgtk/res.go
	go build -ldflags "$(LDFLAGS) $(LDEXTRA)" -o $(GAPP)$(EXT) cmd/fl2kmlgtk/main.go cmd/fl2kmlgtk/res.go

$(_FAPP): $(FSRCS)
	go build -ldflags "$(LDFLAGS) $(LDEXTRA)" -o $(FAPP)$(EXT) cmd/fl2kmlfyne/main.go

$(_QAPP): $(QSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(QAPP)$(EXT) cmd/fl2mqtt/main.go

cmd/fl2kmlgtk/res.go: cmd/fl2kmlgtk/logkml.ui
	tools/packui.sh  cmd/fl2kmlgtk/logkml.ui > cmd/fl2kmlgtk/res.go

clean:
	@rm -f $(CAPP)$(EXT) $(MAPP)$(EXT) $(GAPP)$(EXT) $(FAPP)$(EXT) $(QAPP)$(EXT)
	@go clean

install: $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(prefix)/bin

install-local: $(CAPP) $(MAPP) $(GAPP)
	install -d $(HOME)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(HOME)/bin/
