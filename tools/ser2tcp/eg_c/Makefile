LIBS=
LDFLAGS ?= -s
OPT ?= -O2
prefix ?= $$HOME/.local
STATIC=

XOS := $(shell uname)
ifeq ($(XOS),Linux)
 LIBS += -ludev
endif

SYSTGT := $(shell $(CC) -dumpmachine)
ifneq (, $(findstring -musl, $(SYSTGT)))
	STATIC=-static
endif

ifneq (,$(findstring CYGWIN, $(XOS)))
 EXT=.exe
endif

APP=ser2tcp$(EXT)
STAPP=sertest$(EXT)

CFLAGS=$(OPT) -Wall -Wextra -pedantic -std=gnu17

TARGETS:=$(APP)

ifeq ($(XOS),Darwin)
  CFLAGS+= -target x86_64-apple-macos11
  TARGETS:=universal
  ARMAPP:=$(APP).arm64
  APP:=$(APP).x64
endif



all: $(TARGETS)

$(APP): ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)


$(ARMAPP): ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)  -target arm64-apple-macos11

universal: $(APP) $(ARMAPP)
	lipo -create -output ser2tcp $^


static: ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $(APP) $(STATIC) $^ $(LIBS)

$(STAPP): serial.c find_best.c
	$(CC) $(CFLAGS) -DSERTEST $(LDFLAGS) -o $@ $(STATIC) $^ $(LIBS)

install: $(APP)
	install -d $(prefix)/bin
	install -s -m 755 $(APP) $(prefix)/bin/

clean:
	rm -f *.o *.exe ser2tcp sertest ser2tcp *.swp *~
