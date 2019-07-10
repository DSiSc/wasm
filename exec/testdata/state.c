#include<stdlib.h>
unsigned char *get_state(unsigned char *key, int key_size);
void set_state(unsigned char *key, int key_size, unsigned char *val, int val_size);

void set_global_state(unsigned char *key, int key_size, unsigned char *val, int val_size) {
  set_state(key, key_size, val, val_size);
}

unsigned char *get_global_state(unsigned char *key, int key_size) {
  return  get_state(key, key_size);
}