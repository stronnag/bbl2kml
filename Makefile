APP=bbl2kml
MAPP=mission2kml

SOURCES = $(wildcard cmd/$(APP)/*.go) $(wildcard pkg/*/*.go)
MSRCS = $(wildcard cmd/mission2kml/*.go) $(wildcard pkg/*/*.go)

$(APP): $(SOURCES)
	go build -ldflags "-s -w" -o $(APP) cmd/$(APP)/main.go

$(MAPP): $(MSRCS)
	go build -ldflags "-s -w" -o $(MAPP) cmd/$(MAPP)/main.go

clean:
	@rm -f bbl2kml mission2kml
	@go clean

all: $(APP) $(MAPP)
