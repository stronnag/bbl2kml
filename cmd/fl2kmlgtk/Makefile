
GAPP=fl2kmlgtk
_GAPP=fl2kmlgtk
SRC=main.go resources.c

all: $(_GAPP)

resources.c: resources.xml  logkml.ui logo.svg
	glib-compile-resources --generate-source  --manual-register resources.xml
	glib-compile-resources --generate-header  --manual-register resources.xml

#fl2kmlgtk: $(SRC)
#	go build -ldflags="-s -w"

$(_GAPP): $(SRC)
	go build $(LDF) "$(LDFLAGS) $(LDEXTRA)" -o $(GAPP)$(EXT)

clean:
	rm -f $(GAPP)
