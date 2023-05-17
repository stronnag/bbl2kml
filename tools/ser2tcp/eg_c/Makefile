LIBS=
LDFLAGS ?= -s
OPT ?= -O2
prefix ?= $$HOME/.local

XOS := $(shell uname)
ifeq ($(XOS),Linux)
 LIBS += -ludev
endif

ifneq (,$(findstring CYGWIN, $(XOS)))
 EXT=.exe
endif

APP=ser2tcp$(EXT)

CFLAGS=$(OPT) -Wall -Wextra -pedantic -std=gnu17

all: $(APP)

$(APP): ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)

install: $(APP)
	install -d $(prefix)/bin
	install -s -m 755 $(APP) $(prefix)/bin/

clean:
	rm -f *.o *.exe ser2tcp ~*
