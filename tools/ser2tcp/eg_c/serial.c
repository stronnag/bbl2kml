
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <fcntl.h>
#include <stdbool.h>
#ifdef  __FreeBSD__
# define __BSD_VISIBLE 1
#endif
#include <sys/ioctl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <errno.h>
#ifdef __linux__
#include <asm/termbits.h>
#ifndef TCGETS2
#include <asm-generic/ioctls.h>
#endif
#else
#include <termios.h>
#endif

#ifdef __CYGWIN__
#include <io.h>
#include <windows.h>
#endif

#ifdef __APPLE__
#include <IOKit/serial/ioss.h>
#endif

#include "ser2tcp.h"

#if !defined( __linux__) && !defined(__APPLE__) && !defined(__CYGWIN__)
static int rate_to_constant(int baudrate) {
#ifdef __FreeBSD__
  return baudrate;
#else
#define B(x) case x: return B##x
    switch(baudrate) {
        B(50);     B(75);     B(110);    B(134);    B(150);
        B(200);    B(300);    B(600);    B(1200);   B(1800);
        B(2400);   B(4800);   B(9600);   B(19200);  B(38400);
        B(57600);  B(115200); B(230400);
        default:
          return 0;
    }
#undef B
#endif
}
#endif

static void flush_serial(int fd) {
#ifdef __linux__
  ioctl(fd, TCFLSH, TCIOFLUSH);
#else
  tcflush(fd, TCIOFLUSH);
#endif
}
void close_serial(int fd) {
  flush_serial(fd);
  close(fd);
}

#ifdef __CYGWIN__
static int set_attributes(int fd, serial_opts_t *sopts, int *aspeed) {
  // heresy, but quite a nice API
  HANDLE hdl = (HANDLE)get_osfhandle (fd);
  int res = -1;
  DCB dcb = {0};
  dcb.DCBlength = sizeof(DCB);
  if (GetCommState(hdl, &dcb)) {
    if (sopts->databits == 0) {
      dcb.ByteSize = 8;
    } else {
      dcb.ByteSize = sopts->databits;
    }
    dcb.BaudRate = sopts->baudrate;

    dcb.StopBits = ONESTOPBIT;
    if (sopts->stopbits != NULL && strcmp(sopts->stopbits, "Two") == 0) {
      dcb.StopBits = TWOSTOPBITS;
    }
    if (sopts->parity == NULL || strcmp(sopts->parity, "None") == 0) {
      dcb.Parity = NOPARITY;
    } else {
      if (strcmp(sopts->parity, "Odd")) {
	dcb.Parity = ODDPARITY;
      } else {
	dcb.Parity = EVENPARITY;
      }
    }
    if (SetCommState(hdl, &dcb)) {
      memset(&dcb, 0, sizeof(DCB));
      if (GetCommState(hdl, &dcb)) {
	*aspeed = dcb.BaudRate;
	res = 0;
      }
    }
  }
  return res;
}
#else
static int set_fd_speed(int fd, int rate, int *aspeed) {
  int res = -1;
#ifdef __linux__
  // Just user BOTHER for everything (allows non-standard speeds)
    struct termios2 t;
    if((res = ioctl(fd, TCGETS2, &t)) != -1) {
      t.c_cflag &= ~CBAUD;
      t.c_cflag |= BOTHER;
      t.c_ospeed = t.c_ispeed = rate;
      res = ioctl(fd, TCSETS2, &t);
      if (res != -1) {
	if((res = ioctl(fd, TCGETS2, &t)) != -1) {
	  *aspeed = t.c_ispeed;
	}
      }
    }
#elif __APPLE__
    speed_t speed = rate;
    res = ioctl(fd, IOSSIOSPEED, &speed);
    if (res != -1) {
      struct termios term;
      if (tcgetattr(fd, &term) != -1) {
	*aspeed = cfgetispeed(&term);
      }
    }
#else
  int speed = rate_to_constant(rate);
  struct termios term;
  if (tcgetattr(fd, &term)) return -1;
  if (speed == 0) {
    res = -1;
  } else {
    if(cfsetispeed(&term,speed) != -1) {
      cfsetospeed(&term,speed);
      res = tcsetattr(fd,TCSANOW,&term);
    }
    if(res != -1) {
      memset(&term, 0, sizeof(term));
      res = (tcgetattr(fd, &term));
      if (res != -1) {
	*aspeed = cfgetispeed(&term);
      }
    }
  }
#endif
  return res;
}

static int set_attributes(int fd, serial_opts_t *sopts, int *aspeed) {
  struct termios tio;
  memset (&tio, 0, sizeof(tio));
  int res = -1;
#ifdef __linux__
  res = ioctl(fd, TCGETS, &tio);
#else
  res = tcgetattr(fd, &tio);
#endif
  if (res != -1) {
    // cfmakeraw ...
    tio.c_iflag &= ~(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL | IXON);
    tio.c_oflag &= ~OPOST;
    tio.c_lflag &= ~(ECHO | ECHONL | ICANON | ISIG | IEXTEN);
    tio.c_cflag &= ~(CSIZE | PARENB);
    tio.c_cflag |= CS8;

    tio.c_cc[VTIME] = 0;
    tio.c_cc[VMIN] = 1;

    tio.c_cflag &= ~CSIZE;
    switch (sopts->databits) {
    case 5:
      tio.c_cflag |=  CS5;
      break;
    case 6:
      tio.c_cflag |=  CS6;
      break;
    case 7:
      tio.c_cflag |=  CS7;
      break;
    default:
      tio.c_cflag |=  CS8;
      break;
    }

    tio.c_cflag |=  CREAD|CLOCAL;
    if (sopts->stopbits != NULL && strcmp(sopts->stopbits, "Two") == 0) {
      tio.c_cflag |=  CSTOPB;
    } else {
      tio.c_cflag &=  ~CSTOPB;
    }

    if (sopts->parity == NULL || strcmp(sopts->parity, "None") == 0) {
      tio.c_cflag &= ~PARENB;
    } else {
      tio.c_cflag |= PARENB;
      if (strcmp(sopts->parity, "Odd")) {
	tio.c_cflag |= PARODD;
      } else {
	tio.c_cflag &= ~PARODD;
      }
    }
#ifdef __linux__
    res = ioctl(fd, TCSETS, &tio);
#else
    res = tcsetattr(fd,TCSANOW,&tio);
#endif
  }
  if (res != -1) {
    res = set_fd_speed(fd, sopts->baudrate, aspeed);
  }
  return res;
}
#endif

void report_speed(int rate, int aspeed) {
    if (rate != aspeed) {
      fprintf(stderr, "Warning: device speed %d differs from requested %d\n", aspeed, rate);
    }
}

int open_serial(serial_opts_t *sopts) {
    int fd;
    int aspeed = -1;
    int res = 1;
    fd = open(sopts->devname, O_RDWR|O_NOCTTY);
    if(fd != -1) {
      res = set_attributes(fd, sopts, &aspeed);
      if (res == -1) {
	close(fd);
	fd = -1;
      } else {
	report_speed(sopts->baudrate, aspeed);
      }
    }
    return fd;
}


#ifdef SERTEST
int main(int argc, char **argv) {
  serial_opts_t sopts = {0};
  int brates[] = {57600, 100000, 115200, 200000, 230400, 400000, 420000, 460800, -1};
  int *rates = brates;




  if (argc > 1) {
    if(strncmp(argv[1], DEVBASE, sizeof(DEVBASE)-1) == 0) {
      sopts.devname = argv[1];
    } else {
      find_device_from_desc(argv[1], &sopts.devname);
    }

    if (argc > 2) {
      rates = calloc(argc, sizeof(int));
      int *r = rates;
      for(int j = 2; j < argc; j++) {
	int k = atoi(argv[j]);
	*r++ = k;
      }
      *r = -1;
    }
  }  else {
    sopts.devname = "/dev/ttyUSB0";
  }

  for(int i = 0; ; i++) {
    if (rates[i] == -1) {
      break;
    }
    sopts.baudrate = rates[i];
    int fd = open_serial(&sopts);
    if (fd == -1) {
      fprintf(stderr,"Failed %s : %d\n", sopts.devname, rates[i]);
      break;
    }
    fprintf(stderr,"OK %s : %d\n", sopts.devname, rates[i]);
    close(fd);
  }
  return 0;
}
#endif
