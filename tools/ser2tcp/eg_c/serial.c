
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
#include <termios.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <errno.h>

#ifdef __linux__
#include <linux/serial.h>
#endif

#include "ser2tcp.h"

static int rate_to_constant(int baudrate) {
#define B(x) case x: return B##x
    switch(baudrate) {
        B(50);     B(75);     B(110);    B(134);    B(150);
        B(200);    B(300);    B(600);    B(1200);   B(1800);
        B(2400);   B(4800);   B(9600);   B(19200);  B(38400);
        B(57600);  B(115200); B(230400);
#ifdef __linux__
        B(460800); B(921600);
        B(500000); B(576000); B(1000000); B(1152000); B(1500000);
#endif
#ifdef __FreeBSD__
        B(460800); B(500000);  B(921600);
        B(1000000); B(1500000);
        B(2000000); B(2500000);
        B(3000000); B(3500000);
        B(4000000);
#endif
        default: return 0;
    }
#undef B
}

static void flush_serial(int fd) {
  tcflush(fd, TCIOFLUSH);
}

void close_serial(int fd) {
  flush_serial(fd);
  close(fd);
}

static int set_fd_speed(int fd, int rate) {
  int speed = rate_to_constant(rate);
  struct termios term;
  if (tcgetattr(fd, &term)) return -1;
  if (speed == 0) {
#ifdef __linux__
#include <asm/termios.h>
#include <asm/ioctls.h>
    struct termios2 t;
    int res;
    if((res = ioctl(fd, TCGETS2, &t)) != -1) {
      t.c_ospeed = t.c_ispeed = rate;
      t.c_cflag &= ~CBAUD;
      t.c_cflag |= (BOTHER|CBAUDEX);
      res = ioctl(fd, TCSETS2, &t);
    }
    return res;
#else
    return -1;
#endif
  } else {
    int res = -1;
    if(cfsetispeed(&term,speed) != -1) {
      cfsetospeed(&term,speed);
      res = tcsetattr(fd,TCSANOW,&term);
    }
    return res;
  }
}

int open_serial(serial_opts_t *sopts) {
    int fd;
    fd = open(sopts->devname, O_RDWR|O_NOCTTY);
    if(fd != -1) {
        struct termios tio;
        memset (&tio, 0, sizeof(tio));
        tcgetattr(fd, &tio);
        cfmakeraw(&tio);
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
        tio.c_iflag &= ~IGNBRK;
        tio.c_iflag |=  BRKINT;
        tio.c_iflag &= ~ISTRIP;
        tio.c_iflag &= ~(INLCR | IGNCR | ICRNL);
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
        tio.c_lflag |=  ISIG;
        tio.c_lflag &= ~ICANON;
        tio.c_lflag &= ~(ECHO | ECHOE | ECHOK | ECHONL);
        tcsetattr(fd,TCSANOW,&tio);
        if(set_fd_speed(fd, sopts->baudrate) == -1) {
          close(fd);
          fd = -1;
        }
    }
    return fd;
}
