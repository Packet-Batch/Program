#pragma once
#include <constants.h>

#include <tech/af-xdp/af_xdp.h>

typedef struct sequence__thread_ctx
{
    const char device[MAX_NAME_LEN];
    struct sequence seq;
    u16 seq_cnt;
    struct cli cmd;
    int id;
    struct xsk_socket_info *xsk_info;
    int batch_size;

    u32 src_ip;
    u32 dst_ip;
} sequence__thread_ctx_t;

void sequence__start(const char *interface, struct sequence seq, u16 seqc, struct cli cmd, int batch_size);
void sequence__stop_all(struct config *cfg);