LDFLAGS ?= -s -w
prefix ?= /usr

ifndef DESTDIR
 APP=bbl2kml
 MAPP=mission2kml
else
 APP=$(DESTDIR)/bbl2kml
 MAPP=$(DESTDIR)/mission2kml
endif

ifeq ($(GOOS),windows)
 EXT=.exe
else
 EXT=
endif

SOURCES = $(wildcard cmd/bbl2kml/*.go) $(wildcard pkg/*/*.go)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(wildcard pkg/*/*.go)

$(APP): $(SOURCES)
	go build -ldflags "$(LDFLAGS)" -o $(APP)$(EXT) cmd/bbl2kml/main.go

$(MAPP): $(MSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(MAPP)$(EXT) cmd/mission2kml/main.go

clean:
	@rm -f $(APP)$(EXT) $(MAPP)$(EXT)
	@go clean

install: $(APP) $(MAPP)
	install -d $(prefix)/bin
	install -s $(APP) $(prefix)/bin/bbl2kml
	install -s $(MAPP) $(prefix)/bin/mission2kml

install-local: $(APP) $(MAPP)
	install -d $(HOME)/bin
	install -s $(APP) $(HOME)/bin/bbl2kml
	install -s $(MAPP) $(HOME)/bin/mission2kml

all: $(APP) $(MAPP)
