#pragma once

#include <stdlib.h>
#include <getopt.h>

typedef struct cli_af_xdp
{
    unsigned int queue_set : 1;
    int queue;

    unsigned int wake_up : 1;
    unsigned int shared_umem : 1;
    unsigned short batch_size;
    unsigned int skb_mode : 1;
    unsigned int zero_copy : 1;
    unsigned int copy : 1;
} cli_af_xdp_t;