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

USE_GCCGO ?= 0
ifneq (, $(shell which gccgo 2>/dev/null))
  USE_GCCGO=1
endif

ifeq (, $(USE_GC))
  USE_GCCGO=0
endif

ifeq (0,$(USE_GCCGO))
 LDF=-ldflags
 GOFLAGS += -compiler=gc
else
 GOFLAGS += -compiler=gccgo
 LDEXTRA=-pthread
 LDF=-gccgoflags
endif


export LDFLAGS
export LDEXTRA
export EXT
export LDF
export GOFLAGS

all: $(_CAPP) $(_MAPP) $(_QAPP)

gui: $(_FAPP) $(_GAPP)

PKGCOMMON = $(wildcard pkg/api/*/*.go) $(wildcard pkg/mission/*.go) $(wildcard pkg/geo/*.go) $(wildcard pkg/options/*.go)
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
	CGO_ENABLED=0 go build $(LDF) "$(LDFLAGS) -extldflags -static" -o $(CAPP)$(EXT) cmd/flightlog2kml/main.go

$(_MAPP): $(MSRCS)
	CGO_ENABLED=0 go build $(LDF) "$(LDFLAGS) -extldflags -static" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

$(_GAPP): $(GSRCS)
	make -C cmd/fl2kmlgtk
	mv cmd/fl2kmlgtk/fl2kmlgtk $(GAPP)

$(_FAPP): $(FSRCS)
	go build $(LDF) "$(LDFLAGS) $(LDEXTRA)" -o $(FAPP)$(EXT) cmd/fl2kmlfyne/main.go

$(_QAPP): $(QSRCS)
	CGO_ENABLED=0 go build $(LDF) "$(LDFLAGS) -extldflags -static" -o $(QAPP)$(EXT) cmd/fl2mqtt/main.go

clean:
	@rm -f $(CAPP)$(EXT) $(MAPP)$(EXT) $(GAPP)$(EXT) $(FAPP)$(EXT) $(QAPP)$(EXT)
	make -C  cmd/fl2kmlgtk clean
	@go clean

install: $(CAPP)$(EXE) $(MAPP)$(EXE) $(QAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(QAPP)$(EXE) $(prefix)/bin

install-all: $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(QAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(QAPP)$(EXE) $(prefix)/bin


install-local: $(CAPP) $(MAPP) $(GAPP) $(QAPP)
	install -d $(HOME)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(HOME)/bin/
