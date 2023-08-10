#include <stdbool.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <getopt.h>

#include <sys/socket.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <netinet/tcp.h>

#include <sys/ioctl.h>

#include <errno.h>
#include <libgen.h>

#include "ser2tcp.h"

#define DEVBASE "/dev"

static int lookup_address (char *name, int port, int type, struct sockaddr * addr, socklen_t* len ) {
  struct addrinfo *servinfo, *p;
  struct addrinfo hints = {.ai_family = AF_UNSPEC, .ai_socktype = type, .ai_flags = AI_V4MAPPED|AI_ADDRCONFIG};
  if (name == NULL) {
    hints.ai_flags |= AI_PASSIVE;
  }
  /*
    This nonsense is to uniformly deliver the same sa_family regardless of whether
    name is NULL or non-NULL ** ON LINUX **
    Otherwise, at least on Linux, we get
    - V6,V4 for the non-null case and
    - V4,V6 for the null case, regardless of gai.conf
    Which may confuse consumers
    FreeBSD and Windows behave consistently, giving V6 for Ipv6 enabled stacks
    unless a quad dotted address is specified (or a name resolveds to V4,
    or system policy enforces IPv4 over V6
  */
    struct addrinfo *p4 = NULL;
    struct addrinfo *p6 = NULL;

    int result;
    char aport[16];
    snprintf(aport, sizeof(aport), "%d", port);

    if ((result = getaddrinfo(name, aport, &hints, &servinfo)) != 0) {
      fprintf(stderr, "getaddrinfo: %s\n", gai_strerror(result));
      return result;
    } else {
      for(p = servinfo; p != NULL; p = p->ai_next) {
        if(p->ai_family == AF_INET6)
          p6 = p;
        else if(p->ai_family == AF_INET)
          p4 = p;
      }

      if (p6 != NULL)
        p = p6;
      else if (p4 != NULL)
        p = p4;
      else
        return -1;
      memcpy(addr, p->ai_addr, p->ai_addrlen);
      *len = p->ai_addrlen;
      freeaddrinfo(servinfo);
    }
    return 0;
}

static char * pretty_print_address(struct sockaddr* p) {
  char straddr[INET6_ADDRSTRLEN];
  void *addr;
  uint16_t port;
  if (p->sa_family == AF_INET6) {
    struct sockaddr_in6 * ip = (struct sockaddr_in6*)p;
    addr = &ip->sin6_addr;
    port = ntohs(ip->sin6_port);
  } else {
    struct sockaddr_in * ip = (struct sockaddr_in*)p;
    port = ntohs(ip->sin_port);
    addr = &ip->sin_addr;
  }
  const char *res = inet_ntop(p->sa_family, addr, straddr, sizeof straddr);
  if (res != NULL) {
    int nb = strlen(res)+16;
    char *buf = calloc(nb,1);
    char *ptr = buf;
    if (p->sa_family == AF_INET6) {
      *ptr++='[';
    }
    ptr = stpcpy(ptr, res);
    if (p->sa_family == AF_INET6) {
      *ptr++=']';
    }
    sprintf(ptr, ":%d", port);
    return buf;
  }
  return NULL;
}

static void usage (char *pname) {
  char* xpname = strdup(pname);
  basename(xpname);
  fprintf(stderr,"%s [options]\n", xpname);
  free(xpname);
  fprintf(stderr, "Options:\n"
          "    -h, --help              print this help menu\n"
          "    -V, --version           print version and exit\n"
          "    -v, --verbose           print I/O read sizes\n"
          "    -c, --comport <name>    serial device name (mandatory)\n"
          "    -b, --baudrate <115200> serial baud rate\n"
          "    -d, --databits <8>      serial databits 5|6|7|8\n"
          "    -s, --stopbits <One>    serial stopbits [None|One|Two]\n"
          "    -p, --parity <None>     serial parity [Even|None|Odd]\n"
          "    -i, --ip <localhost>    Host name / Address\n"
          "    -t, --tcpport <5762>    IP port\n"
          "    -z, --buffersize <n>    Buffersize (ignored)\n");
  exit(1);
}

static void version(void) {
  fprintf(stderr,"%s\n", VERSION);
  exit(0);
}

int main(int argc, char **argv) {
  struct option longOpt[] = {
    {"baudrate", required_argument, 0, 'b'},
    {"comport", required_argument, 0, 'c'},
    {"databits", required_argument, 0, 'd'},
    {"ip", required_argument, 0, 'i'},
    {"parity", required_argument, 0, 'p'},
    {"stopbits", required_argument, 0, 's'},
    {"tcpport", required_argument, 0, 't'},
    {"buffersize", required_argument, 0, 'z'},
    {"version", no_argument, 0, 'V'},
    {"verbose", no_argument, 0, 'v'},
    {"help", no_argument, 0, 'h'},
    {NULL, 0, NULL, 0}
  };

  serial_opts_t seropts = {.baudrate = 115200};

  char *host = "localhost";
  int port = 5762;
  bool verbose = false;
  char *devnode = NULL;

  int c;
  for (bool done = false; !done; ) {
    c = getopt_long(argc, argv, "hVvc:b:d:s:p:i:t:z:", longOpt, NULL);
    switch (c) {
    case -1:
      done = true;
      break;
    case 'b':
      seropts.baudrate = atoi(optarg);
      break;
    case 'c':
      devnode = strdup(optarg);
      break;
    case 'd':
      seropts.databits = atoi(optarg);
      break;
    case 'i':
      host = strdup(optarg);
      break;
    case 'p':
      seropts.parity = strdup(optarg);
      break;
    case 's':
      seropts.stopbits = strdup(optarg);
      break;
    case 't':
      port = atoi(optarg);
      break;
    case 'V':
      version();
      break;
    case 'v':
      verbose = true;
      break;
    case 'z':
      break;
    default:
      usage(argv[0]);
      break;
    }
  }

  if (devnode == NULL) {
    usage(argv[0]);
  }

  struct sockaddr_storage saddr;
  static socklen_t saddr_len;

  if(lookup_address(host, port, SOCK_STREAM, (struct sockaddr*)&saddr, &saddr_len) != 0) {
    fprintf(stderr, "Failed to resolve %s %d\n", host, port);
    return 127;
    }

  int sockfd = socket(((struct sockaddr*)&saddr)->sa_family, SOCK_STREAM, IPPROTO_TCP);
  if (sockfd == -1) {
    fprintf(stderr, "Failed to open socket for %s %d\n", host, port);
    return 127;
  }

  int res = -1;
  for(int j = 0; j < 20; j++) {
    res = connect(sockfd, (struct sockaddr *)&saddr, saddr_len);
    if (res == 0) {
      break;
    }
    usleep(250*1000);
  }
  if (res == -1) {
    fprintf(stderr, "Failed to connect to %s %d\n", host, port);
    return 127;
  }

  int one = 1;
  setsockopt(sockfd, IPPROTO_TCP, TCP_NODELAY, &one, sizeof(one));

  if(strncmp(devnode, DEVBASE, sizeof(DEVBASE)-1) == 0) {
    seropts.devname = devnode;
  } else {
    find_device_from_desc(devnode, &seropts.devname);
    free(devnode);
  }

  int fd = open_serial(&seropts);
  if (fd == -1) {
    fprintf(stderr,"Failed to open %s\n", seropts.devname);
    return 127;
  } else {
    char *pretty = pretty_print_address((struct sockaddr *)&saddr);
    if (pretty != NULL) {
      fprintf(stderr,"Connected to %s <-> %s\n", pretty, seropts.devname);
      free(pretty);
    }
  }

  fd_set sfds,rfds;
  char buf[1024];
  FD_ZERO (&sfds);
  FD_SET (sockfd, &sfds);
  FD_SET (fd, &sfds);
  for(bool done = false;!done;) {
    int sts;
    int n, nb;
    rfds = sfds;
    sts = select (FD_SETSIZE, &rfds, NULL, NULL, NULL);
    switch(sts) {
      case -1:
        done = true;
        break;
    case 0:
      /* Should never happen */
      done = true;
      break;
    default:
      if (FD_ISSET (sockfd, &rfds)) {
        n = ioctl(sockfd, FIONREAD, &nb);
        if (n == 0) {
          if (nb > 1024) {
            nb = 1024;
          }
          int res = recv(sockfd, buf, nb, 0);
          if(verbose) {
            fprintf(stderr,"read %d from socket\n", res);
          }
          if (res > 0) {
            res = write(fd, buf, res);
            if (verbose)
              fprintf(stderr,"write %d to serial\n", res);
          }
          if (res <= 0) {
            done = true;
          }
        } else {
          done = true;
        }
      }

      if (FD_ISSET (fd, &rfds)) {
        int res = read(fd, buf, 1024);
        if(verbose) {
          fprintf(stderr,"read %d from serial\n", res);
        }
        if (res > 0) {
          res = send(sockfd, buf, res, 0);
            if (verbose)
              fprintf(stderr,"send %d to socket\n", res);
        }
        if (res <= 0) {
          done = true;
        }
      }
    }
  }
  return 0;
}
