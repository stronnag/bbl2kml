LDFLAGS=-s
LIBS=
OPT ?= -O2
prefix ?= $$HOME/.local

XOS := $(shell uname)

ifeq ($(XOS),Linux)
 LIBS=-ludev
endif

CFLAGS=$(OPT) -Wall

all:	ser2tcp

ser2tcp: ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)

install: ser2tcp
	install -d $(prefix)/bin
	install -s -m 755 ser2tcp $(prefix)/bin/

clean:
	rm -f *.o *.exe ser2tcp ~*
