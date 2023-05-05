#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <stdbool.h>

#ifdef __linux__
#include <libudev.h>

int find_device_from_desc(char* desc, char**pdev) {
  struct udev *udev;
  struct udev_enumerate *enumerate;
  struct udev_list_entry *devices, *dev_list_entry;
  struct udev_device *dev;

  udev = udev_new();
  if (!udev) {
    return -1;
  }
  enumerate = udev_enumerate_new(udev);
  udev_enumerate_add_match_subsystem(enumerate, "tty");
  udev_enumerate_scan_devices(enumerate);
  devices = udev_enumerate_get_list_entry(enumerate);
  bool found = false;
  udev_list_entry_foreach(dev_list_entry, devices) {
    const char *path;
    path = udev_list_entry_get_name(dev_list_entry);
    dev = udev_device_new_from_syspath(udev, path);
    const char *devnode = udev_device_get_devnode(dev);
    /* for details, we need to search up the tree */
    dev = udev_device_get_parent_with_subsystem_devtype(
		       dev,
		       "usb",
		       "usb_device");
    if (dev) {
      const char*product = udev_device_get_sysattr_value(dev,"product");
      if (product != NULL && strcmp(product, desc) == 0) {
        if(pdev != NULL) {
          *pdev = strdup(devnode);
        }
        found = true;
      }
    }
    udev_device_unref(dev);
    if (found)
      break;
  }
  udev_enumerate_unref(enumerate);
  udev_unref(udev);
  if (!found) {
    char usbdev[16];
    for(int i = 0; i < 9; i++) {
      sprintf(usbdev,"/dev/ttyUSB%d", i);
      if (access(usbdev, R_OK|W_OK) == 0) {
        found = true;
        if(pdev != NULL) {
          *pdev = strdup(usbdev);
        }
        break;
      }
    }
  }
  return (found) ? 0 : 1;
}
#elif __FreeBSD__
int find_device_from_desc(char* desc __attribute__((unused)), char**pdev) {
  char usbdev[16];
  for(int i = 0; i < 9; i++) {
    sprintf(usbdev,"/dev/cuaU%d", i);
    if (access(usbdev, R_OK|W_OK) == 0) {
      if(pdev != NULL) {
        *pdev = strdup(usbdev);
      }
      break;
    }
  }
  return 0;
}
#elif __APPLE__
int find_device_from_desc(char* desc  __attribute__((unused)), char**pdev) {
  char usbdev[32];
  for(int i = 0; i < 9; i++) {
    sprintf(usbdev,"/dev/cu.usbserial-000%d", i);
    if (access(usbdev, R_OK|W_OK) == 0) {
      if(pdev != NULL) {
        *pdev = strdup(usbdev);
      }
      break;
    }
  }
  return 0;
}
#elif __CYGWIN__
int find_device_from_desc(char* desc, char**pdev) {
  int dn;
  if (sscanf(desc, "COM%d", &dn) == 1) {
    *pdev = malloc(32);
    sprintf(*pdev, "/dev/ttyS%d", dn-1);
    return 0;
  }
  return -1;
}
#else
int find_device_from_desc(char* desc, char**pdev) {
  return -1;
}
#endif
