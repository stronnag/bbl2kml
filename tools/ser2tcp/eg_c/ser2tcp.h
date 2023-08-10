
#define VERSION "0.0.2"

typedef struct {
  int baudrate;
  int databits;
  char *stopbits;
  char *parity;
  char *devname;
} serial_opts_t;

extern void close_serial(int);
extern int open_serial(serial_opts_t *);
extern int find_device_from_desc(char*, char**);
