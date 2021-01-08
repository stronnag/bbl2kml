LDFLAGS ?= -s -w
prefix ?= /usr

ifndef DESTDIR
 APP=bbl2kml
 MAPP=mission2kml
 OAPP=otx2kml
else
 APP=$(DESTDIR)/bbl2kml
 MAPP=$(DESTDIR)/mission2kml
 OAPP=$(DESTDIR)/otx2kml
endif

ifeq ($(GOOS),windows)
 EXT=.exe
else
 EXT=
endif

all: $(APP)$(EXT) $(MAPP)$(EXT) $(OAPP)$(EXT)

PKGCOMMON = $(wildcard pkg/api/*/*.go) $(wildcard pkg/mission/*.go) $(wildcard pkg/geo/*.go)
PKGOPT = $(wildcard pkg/options/*.go)
PKGBBL = $(wildcard pkg/bbl/*.go)
PKGOTX = $(wildcard pkg/otx/*.go)
PKGINAV = $(wildcard pkg/inav/*.go)
PKGKML = $(wildcard pkg/kmlgen/*.go)

ASRCS = $(wildcard cmd/bbl2kml/*.go) $(PKGCOMMON) $(PKGBBL) $(PKGINAV) $(PKGKML)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(PKGCOMMON)
OSRCS = $(wildcard cmd/otx2kml/*.go) $(PKGCOMMON) $(PKGOTX) $(PKGKML)

$(APP)$(EXT): $(ASRCS)
	go build -ldflags "$(LDFLAGS)" -o $(APP)$(EXT) cmd/bbl2kml/main.go

$(MAPP)$(EXT): $(MSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

$(OAPP)$(EXT): $(OSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(OAPP)$(EXT) cmd/otx2kml/main.go

clean:
	@rm -f $(APP)$(EXT) $(MAPP)$(EXT) $(OAPP)$(EXT)
	@go clean

install: $(APP)$(EXE) $(MAPP)$(EXE) $(OAPP)$(EXE) $(OAPP)$(EXE)
	@ install -d $(prefix)/bin
	install -s $(APP)$(EXE) $(MAPP)$(EXE) $(OAPP)$(EXE) $(OAPP)$(EXE) $(prefix)/bin/

install-local: $(APP) $(MAPP)
	install -d $(HOME)/bin
	install -s $(APP)$(EXE) $(MAPP)$(EXE) $(OAPP)$(EXE) $(OAPP)$(EXE) $(HOME)/bin/
