#include<stdlib.h>
char* get_str() {
	char *str =(char*)malloc(5);
	str[0] = 'h';
	str[1] = 'e';
	str[2] = 'l';
	str[3] = 'l';
	str[4] = 'o';
	return str;
}

void reverse_str(char *str, int size) {
  char tmp;
  for (int i = 0; i < size/2; i++) {
    tmp = str[i];
    str[i] = str[size - i -1];
    str[size - i - 1] = tmp;
  }
}