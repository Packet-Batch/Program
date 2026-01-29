#pragma once

#include <helpers/int_types.h>

void get_gw_mac(u8 *mac);
int get_src_mac_address(const char *dev, u8 *src_mac);
u16 rand_num(u16 min, u16 max, unsigned int seed);
char *lower_str(char *str);
char *rand_ip(char *range, unsigned int seed);