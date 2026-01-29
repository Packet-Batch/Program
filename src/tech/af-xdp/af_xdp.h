#pragma once

#include <xsk.h>

#include <tech/af-xdp/cli.h>

#define MAX_CPUS 256
#define NUM_FRAMES 4096
#define FRAME_SIZE XSK_UMEM__DEFAULT_FRAME_SIZE
#define INVALID_UMEM_FRAME UINT64_MAX
// #define DEBUG

#define MAX_BATCH_SIZE 4096
#define MAX_PKT_LEN 0xFFFF

typedef struct xsk_umem_info
{
    struct xsk_ring_prod fq;
    struct xsk_ring_cons cq;
    struct xsk_umem *umem;
    void *buffer;
} xsk_umem_info_t;

typedef struct xsk_socket
{
    struct xsk_ring_cons *rx;
    struct xsk_ring_prod *tx;
    u64 outstanding_tx;
    struct xsk_ctx *ctx;
    struct xsk_socket_config config;
    int fd;
} xsk_socket_t;

typedef struct xsk_socket_info
{
    struct xsk_ring_cons rx;
    struct xsk_ring_prod tx;
    struct xsk_umem_info *umem;
    struct xsk_socket *xsk;

    u64 umem_frame_addr[NUM_FRAMES];
    u32 umem_frame_free;

    u32 outstanding_tx;
} xsk_socket_info_t;

int send_packet(struct xsk_socket_info *xsk, int thread_id, void *pckt, u16 length, u8 verbose);
int send_packet_batch(struct xsk_socket_info *xsk, void *pkts, u16 *pkts_len, u16 amt);
void flush_send_queue(struct xsk_socket_info *xsk);
u64 get_umem_addr(struct xsk_socket_info *xsk, int idx);
void *get_umem_loc(struct xsk_socket_info *xsk, u64 addr);
void setup_af_xdp_variables(struct cli_af_xdp *cmd_af_xdp, int verbose);
struct xsk_umem_info *setup_umem(int index);
struct xsk_socket_info *setup_socket(const char *dev, u16 thread_id, int verbose);
void cleanup_socket(struct xsk_socket_info *xsk);
int get_socket_fd(struct xsk_socket_info *xsk);