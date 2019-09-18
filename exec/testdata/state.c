#include<stdlib.h>
unsigned char *get_state(unsigned char *key);
void set_state(unsigned char *key, unsigned char *val);

void set_global_state(unsigned char *key, unsigned char *val) {
  set_state(key, val);
}

unsigned char *get_global_state(unsigned char *key) {
  return  get_state(key);
}