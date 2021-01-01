APP = bbl2kml

prefix?=$(DESTDIR)/usr

$(APP):
	go build -ldflags "-w -s"

install: $(APP)
	install -d $(prefix)/bin
	install -s $(APP) $(prefix)/bin/$(APP)

clean:
	go clean
