#pragma once

#include <helpers/int_types.h>

void utils__get_gw_mac(u8 *mac);
int utils__get_src_mac_addr(const char *dev, u8 *src_mac);
u16 utils__rand_num(u16 min, u16 max, unsigned int seed);
char *utils__lower_str(char *str);
char *utils__rand_ip(char *range, unsigned int seed);