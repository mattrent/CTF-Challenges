#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main() {
  const char flag[] = {102, 107, 95, 100, 119, 110, 110, 90, 108, 96,
                       89,  34,  89, 107, 83,  94,  96,  91, 83,  106};
  const size_t length = sizeof(flag) / sizeof(flag[0]);
  char input[256] = {0};

  printf("Enter the flag:\n");
  if (fgets(input, sizeof(input), stdin) == NULL) {
    printf("Input missing\n");
    return 1;
  }

  int value = 0;
  for (size_t i = 0; i < length; i++) {
    value += abs(input[i] - (flag[i] + (int)i));
  }

  if (value != 0) {
    printf("Wrong flag\n");
    return 1;
  } else {
    printf("Correct\n");
    return 0;
  }
}
