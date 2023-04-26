LDFLAGS=-s
LIBS=
OPT=-O2

XOS := $(shell uname)

ifeq ($(XOS),Linux)
 LIBS=-ludev
endif

CFLAGS=$(OPT) -Wall

all:	ser2tcp

ser2tcp: ser2tcp.c serial.c find_best.c
	$(CC) $(CFLAGS) $(LDFLAGS) -o $@ $^ $(LIBS)

clean:
	rm -f *.o *.exe ser2tcp ~*
