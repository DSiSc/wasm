#include<stdlib.h>
unsigned char *justitia_internal_storage_read(unsigned char *key);
void justitia_internal_storage_write(unsigned char *key, unsigned char *val);

void set_global_state(unsigned char *key, unsigned char *val) {
  justitia_internal_storage_write(key, val);
}

unsigned char *get_global_state(unsigned char *key) {
  return  justitia_internal_storage_read(key);
}