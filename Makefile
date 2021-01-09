LDFLAGS ?= -s -w
prefix ?= /usr

ifndef DESTDIR
 FAPP=flightlog2kml
 MAPP=mission2kml
else
 FAPP=flightlog2kml
 MAPP=$(DESTDIR)/mission2kml
endif

ifeq ($(GOOS),windows)
 EXT=.exe
else
 EXT=
endif

all: $(FAPP)$(EXT) $(BAPP)$(EXT) $(MAPP)$(EXT) $(OAPP)$(EXT)

PKGCOMMON = $(wildcard pkg/api/*/*.go) $(wildcard pkg/mission/*.go) $(wildcard pkg/geo/*.go)
PKGOPT = $(wildcard pkg/options/*.go)
PKGBBL = $(wildcard pkg/bbl/*.go)
PKGOTX = $(wildcard pkg/otx/*.go)
PKGINAV = $(wildcard pkg/inav/*.go)
PKGKML = $(wildcard pkg/kmlgen/*.go)

FSRCS = $(wildcard cmd/flightlog2kml/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGOTX) $(PKGINAV) $(PKGKML)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(PKGCOMMON)

$(FAPP)$(EXT): $(FSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(FAPP)$(EXT) cmd/flightlog2kml/main.go

$(MAPP)$(EXT): $(MSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

clean:
	@rm -f $(FAPP)$(EXT) $(MAPP)$(EXT)
	@go clean

install: $(FAPP)$(EXE) $(MAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(FAPP)$(EXE) $(MAPP)$(EXE) $(prefix)/bin

install-local: $(FAPP) $(MAPP)
	install -d $(HOME)/bin
	install -s $(FAPP)$(EXE) $(MAPP)$(EXE) $(HOME)/bin/
