DESTDIR ?= .
LDFLAGS ?= -s -w
prefix ?= /usr

APP=$(DESTDIR)/bbl2kml
MAPP=$(DESTDIR)/mission2kml


SOURCES = $(wildcard cmd/$(APP)/*.go) $(wildcard pkg/*/*.go)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(wildcard pkg/*/*.go)

$(APP): $(SOURCES)
	go build -ldflags "$(LDFLAGS)" -o $(APP) cmd/bbl2kml/main.go

$(MAPP): $(MSRCS)
	go build -ldflags "$(LDFLAGS)" -o $(MAPP) cmd/mission2kml/main.go

clean:
	@rm -f $(APP) $(MAPP)
	@go clean

install: $(APP) $(MAPP)
	install -d $(prefix)/bin
	install -s $(APP) $(prefix)/bin/bbl2kml
	install -s $(MAPP) $(prefix)/bin/missoin2kml

install-local: $(APP) $(MAPP)
	install -d $(HOME)/bin
	install -s $(APP) $(HOME)/bin/bbl2kml
	install -s $(MAPP) $(HOME)/bin/missoin2kml

all: $(APP) $(MAPP)
