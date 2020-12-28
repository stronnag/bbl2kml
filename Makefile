
APP = bbl2kml

APP:
	go build -ldflags "-w -s"

install: $(APP)
	install -d $(prefix)/bin
	install -s $(APP) $(prefix)/bin/$(APP)

clean:
	go clean
