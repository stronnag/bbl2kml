LDFLAGS ?= -s -w
prefix ?= /usr

_CAPP=flightlog2kml
_MAPP=mission2kml
_QAPP=fl2mqtt
_LAPP=log2mission

ifndef DESTDIR
 CAPP=flightlog2kml
 MAPP=mission2kml
 FAPP=fl2kmlfyne
 QAPP=fl2mqtt
 LAPP=log2mission
else
 CAPP=$(DESTDIR)/flightlog2kml
 MAPP=$(DESTDIR)/mission2kml
 QAPP=$(DESTDIR)/fl2mqtt
 LAPP=$(DESTDIR)/log2mission
endif

ifeq ($(GOOS),windows)
 EXT=.exe
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

all: $(_CAPP) $(_MAPP) $(_QAPP) $(_LAPP)

PKGCOMMON = $(wildcard pkg/api/*/*.go) $(wildcard pkg/mission/*.go) $(wildcard pkg/geo/*.go) $(wildcard pkg/options/*.go)
PKGBBL = $(wildcard pkg/bbl/*.go)
PKGOTX = $(wildcard pkg/otx/*.go)
PKGINAV = $(wildcard pkg/inav/*.go)
PKGKML = $(wildcard pkg/kmlgen/*.go)
PKGMQTT = $(wildcard pkg/bltmqtt/*.go)
PKGLTM = $(wildcard pkg/ltmgen/*.go)
PKGBLTR = $(wildcard pkg/bltreader/*.go)
PKGL2M = $(wildcard pkg/log2mission/*.go)

CSRCS = $(wildcard cmd/flightlog2kml/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)  $(PKGBLTR)
QSRCS = $(wildcard cmd/fl2mqtt/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGMQTT) $(PKGLTM)
#LSRCS = $(wildcard cmd/fl2ltm/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGLTM)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(PKGCOMMON)
LSRCS = cmd/log2mission/main.go $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)  $(PKGBLTR) $(PKGL2M)

$(_CAPP): $(CSRCS)
	CGO_ENABLED=0 go build -trimpath $(LDF) "$(LDFLAGS) -extldflags -static" -o $(CAPP)$(EXT) cmd/flightlog2kml/main.go

$(_LAPP): $(LSRCS)
	CGO_ENABLED=0 go build -trimpath $(LDF) "$(LDFLAGS) -extldflags -static" -o $(LAPP)$(EXT) cmd/log2mission/main.go

$(_MAPP): $(MSRCS)
	CGO_ENABLED=0 go build -trimpath $(LDF) "$(LDFLAGS) -extldflags -static" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

$(_QAPP): $(QSRCS)
	CGO_ENABLED=0 go build -trimpath $(LDF) "$(LDFLAGS) -extldflags -static" -o $(QAPP)$(EXT) cmd/fl2mqtt/main.go
	ln -sf fl2mqtt fl2ltm

clean:
	@rm -f $(CAPP)$(EXT) $(MAPP)$(EXT) $(GAPP)$(EXT) $(FAPP)$(EXT) $(QAPP)$(EXT) fl2ltm* $(LAPP)$(EXT)
	@go clean

install: $(CAPP)$(EXE) $(MAPP)$(EXE) $(QAPP)$(EXE) $(LAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(QAPP)$(EXE) $(LAPP)$(EXE) $(prefix)/bin
	@rm -f $(prefix)/bin/fl2ltm
	@ln -sf $(prefix)/bin/fl2mqtt $(prefix)/bin/fl2ltm

install-all: $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(QAPP)$(EXE) $(LAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(QAPP)$(EXE) $(LAPP)$(EXE) $(prefix)/bin


install-local: $(CAPP) $(MAPP) $(GAPP) $(QAPP) $(LAPP)
	install -d $(HOME)/bin
	install -s $(CAPP)$(EXE) $(MAPP)$(EXE) $(GAPP)$(EXE) $(LAPP)$(EXE) $(HOME)/bin/
