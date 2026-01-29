#pragma once
#include <constants.h>

#include <tech/af-xdp/af_xdp.h>

// #define VERY_RANDOM

#define MAX_PCKT_LEN 0xFFFF
#define MAX_THREADS 4096
#define MAX_NAME_LEN 64

typedef struct thread_info
{
    const char device[MAX_NAME_LEN];
    struct sequence seq;
    u16 seq_cnt;
    struct cmd_line cmd;
    int id;
    struct xsk_socket_info *xsk_info;
    int batch_size;

    u32 src_ip;
    u32 dst_ip;
} thread_info_t;

void seq_send(const char *interface, struct sequence seq, u16 seqc, struct cmd_line cmd, int batch_size);
void shutdown_prog(struct config *cfg);