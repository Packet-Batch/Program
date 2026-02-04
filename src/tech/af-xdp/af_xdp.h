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

int tech_afxdp__send_pkts(struct xsk_socket_info *xsk, void *pkts, u16 *pkts_len, u16 amt);
void tech_afxdp__setup_vars(struct cli_af_xdp *cmd_af_xdp, int verbose);
struct xsk_socket_info *tech_afxdp__sock_setup(const char *dev, u16 thread_id, int verbose);
void tech_afxdp__sock_cleanup(struct xsk_socket_info *xsk);
int tech_afxdp__sock_fd(struct xsk_socket_info *xsk);