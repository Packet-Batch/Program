#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <locale.h>

#include <net/if.h>

#include <sys/socket.h>
#include <linux/if_link.h>
#include <bpf.h>

#include <helpers/int_types.h>

#include <tech/af-xdp/af_xdp.h>

#include <constants.h>

/* Global variables */
// The XDP flags to load the AF_XDP/XSK sockets with.
u32 xdp_flags = XDP_FLAGS_DRV_MODE;
u32 bind_flags = XDP_USE_NEED_WAKEUP;
int is_shared_umem = 0;
u16 batch_size = 64;
int static_queue_id = 0;
int queue_id = 0;

// For shared UMEM.
static unsigned int global_frame_idx = 0;

// Pointers to the umem and XSK sockets for each thread.
xsk_umem_info_t *shared_umem = NULL;

#ifdef AF_XDP_NO_REFILL_FRAMES
int cur_frame_idx = 0;
#endif

/**
 * Allocates a UMEM frame by retrieving it from the free list.
 *
 * @param xsk A pointer to the xsk_socket_info structure.
 *
 * @return The allocated frame address or INVALID_UMEM_FRAME on failure.
 **/
static inline u64 xsk_alloc_umem_frame(struct xsk_socket_info *xsk)
{
    u64 frame;

#ifdef AF_XDP_NO_REFILL_FRAMES
    if (xsk->umem_frame_free < 1)
    {
        frame = xsk->umem_frame_addr[cur_frame_idx++];

        if (cur_frame_idx >= NUM_FRAMES)
            cur_frame_idx = 0;

        return frame;
    }
#endif

    if (xsk->umem_frame_free == 0)
        return INVALID_UMEM_FRAME;

    frame = xsk->umem_frame_addr[--xsk->umem_frame_free];
    xsk->umem_frame_addr[xsk->umem_frame_free] = INVALID_UMEM_FRAME;

    return frame;
}

/**
 * Frees a UMEM frame by adding it back to the free list.
 *
 * @param xsk A pointer to the xsk_socket_info structure.
 * @param frame The frame address to free.
 *
 * @return Void
 **/
static void xsk_free_umem_frame(struct xsk_socket_info *xsk, u64 frame)
{
    if (xsk->umem_frame_free >= NUM_FRAMES)
    {
        fprintf(stderr, "UMEM frame free list is already full. Cannot free more frames.\n");

        return;
    }

    xsk->umem_frame_addr[xsk->umem_frame_free++] = frame;
}

/**
 * Completes the TX call via a syscall and also checks if we need to free the TX buffer.
 *
 * @param xsk A pointer to the xsk_socket_info structure.
 *
 * @return Void
 **/
static void complete_tx(xsk_socket_info_t *xsk)
{
    // Initiate starting variables (completed amount and completion ring index).
    unsigned int completed;
    uint32_t idx_cq;

    // If outstanding is below 1, it means we have no packets to TX.
    if (!xsk->outstanding_tx)
    {
        return;
    }

    // If we need to wakeup, execute syscall to wake up socket.
    if (!(bind_flags & XDP_USE_NEED_WAKEUP) || xsk_ring_prod__needs_wakeup(&xsk->tx))
    {
        sendto(xsk_socket__fd(xsk->xsk), NULL, 0, MSG_DONTWAIT, NULL, 0);
    }

    // Try to free n (batch_size) frames on the completetion ring.
    completed = xsk_ring_cons__peek(&xsk->umem->cq, batch_size, &idx_cq);

    if (completed > 0)
    {
        for (int i = 0; i < completed; i++)
        {
            xsk_free_umem_frame(xsk, *xsk_ring_cons__comp_addr(&xsk->umem->cq, idx_cq++));
        }

        // Release "completed" frames.
        xsk_ring_cons__release(&xsk->umem->cq, completed);

        xsk->outstanding_tx -= completed < xsk->outstanding_tx ? completed : xsk->outstanding_tx;
    }
}

/**
 * Configures the UMEM area for our AF_XDP/XSK sockets to use for rings.
 *
 * @param buffer The blank buffer we allocated in setup_socket().
 * @param size The buffer size.
 *
 * @return Returns a pointer to the UMEM area instead of the XSK UMEM information structure (struct xsk_umem_info).
 **/
static xsk_umem_info_t *configure_xsk_umem(void *buffer, u64 size)
{
    // Create umem pointer and return variable.
    xsk_umem_info_t *umem;
    int ret;

    // Allocate memory space to the umem pointer and check.
    umem = calloc(1, sizeof(*umem));

    if (!umem)
    {
        return NULL;
    }

    // Attempt to create the umem area and check.
    ret = xsk_umem__create(&umem->umem, buffer, size, &umem->fq, &umem->cq, NULL);

    if (ret)
    {
        errno = -ret;

        return NULL;
    }

    // Assign the buffer we created in setup_socket() to umem buffer.
    umem->buffer = buffer;

    // Return umem pointer.
    return umem;
}

/**
 * Configures an AF_XDP/XSK socket.
 *
 * @param umem A pointer to the umem we created in setup_socket().
 * @param queue_id The TX queue ID to use.
 * @param dev The name of the interface we're binding to.
 *
 * @return Returns a pointer to the AF_XDP/XSK socket inside of a the XSK socket info structure (struct xsk_socket_info).
 **/
static xsk_socket_info_t *xsk_configure_socket(xsk_umem_info_t *umem, int queue_id, const char *dev)
{
    // Initialize starting variables.
    struct xsk_socket_config xsk_cfg;
    struct xsk_socket_info *xsk_info;
    u32 idx;
    int i;
    int ret;

    // Allocate memory space to our XSK socket.
    xsk_info = calloc(1, sizeof(*xsk_info));

    // If it fails, return.
    if (!xsk_info)
    {
        fprintf(stderr, "Failed to allocate memory space to AF_XDP/XSK socket.\n");

        return NULL;
    }

    // Assign AF_XDP/XSK's socket umem area to the umem we allocated before.
    xsk_info->umem = umem;

    // Set the TX size (we don't need anything RX-related).
    xsk_cfg.tx_size = XSK_RING_PROD__DEFAULT_NUM_DESCS;

    // Make sure we don't load an XDP program via LibBPF.
    xsk_cfg.libbpf_flags = XSK_LIBBPF_FLAGS__INHIBIT_PROG_LOAD;

    // Assign our XDP flags.
    xsk_cfg.xdp_flags = xdp_flags;

    // Assign bind flags.
    xsk_cfg.bind_flags = bind_flags;

    // Attempt to create the AF_XDP/XSK socket itself at queue ID (we don't allocate a RX queue for obvious reasons).
    ret = xsk_socket__create(&xsk_info->xsk, dev, queue_id, umem->umem, NULL, &xsk_info->tx, &xsk_cfg);

    if (ret)
    {
        fprintf(stderr, "Failed to create AF_XDP/XSK socket at creation.\n");

        goto error_exit;
    }

    // Assign each umem frame to an address we'll use later.
    for (i = 0; i < NUM_FRAMES; i++)
    {
        xsk_info->umem_frame_addr[i] = i * FRAME_SIZE;
    }

    // Assign how many number of frames we can hold.
    xsk_info->umem_frame_free = NUM_FRAMES;

    // Stuff the receive path.
    ret = xsk_ring_prod__reserve(&xsk_info->umem->fq, XSK_RING_PROD__DEFAULT_NUM_DESCS, &idx);

    if (ret != XSK_RING_PROD__DEFAULT_NUM_DESCS)
    {
        fprintf(stderr, "Failed to reserve fill ring for UMEM.\n");

        goto error_exit;
    }

    for (i = 0; i < XSK_RING_PROD__DEFAULT_NUM_DESCS; i++)
    {
        *xsk_ring_prod__fill_addr(&xsk_info->umem->fq, idx++) = xsk_alloc_umem_frame(xsk_info);
    }

    xsk_ring_prod__submit(&xsk_info->umem->fq, XSK_RING_PROD__DEFAULT_NUM_DESCS);

    // Return the AF_XDP/XSK socket information itself as a pointer.
    return xsk_info;

// Handle error and return NULL.
error_exit:
    errno = -ret;

    return NULL;
}

/**
 * Internal function to send a batch of packets.
 *
 * @param xsk A pointer to the XSK socket info.
 * @param packets Array of packet buffers.
 * @param lengths Array of packet lengths.
 * @param amt Number of packets to send.
 *
 * @return Returns 0 on success and -1 on failure.
 **/
static inline int send_packet_batch_internal(xsk_socket_info_t *xsk, void *buffer, u16 *lengths, u16 amt)
{
    u32 tx_idx = 0;

    while (xsk_ring_prod__reserve(&xsk->tx, amt, &tx_idx) < amt)
    {
        complete_tx(xsk);
    }

    u8 *buf_ptr = (u8 *)buffer;

    for (int i = 0; i < amt; i++)
    {
        u64 frame = xsk_alloc_umem_frame(xsk);

        if (frame == INVALID_UMEM_FRAME)
        {
            fprintf(stderr, "Failed to allocate UMEM frame\n");

            return -1;
        }
#ifdef AF_XDP_NO_REFILL_FRAMES
        if (xsk->umem_frame_free > 0)
        {
#endif
            memcpy(get_umem_loc(xsk, frame), buf_ptr + (i * MAX_PKT_LEN), lengths[i]);
#ifdef AF_XDP_NO_REFILL_FRAMES
        }
#endif

        struct xdp_desc *tx_desc = xsk_ring_prod__tx_desc(&xsk->tx, tx_idx + i);
        tx_desc->addr = frame;
        tx_desc->len = lengths[i];
    }

    xsk_ring_prod__submit(&xsk->tx, amt);
    xsk->outstanding_tx += amt;

    if (xsk->outstanding_tx >= batch_size / 2)
    {
        complete_tx(xsk);
    }

    return 0;
}

/**
 * Sends a single packet buffer out the AF_XDP socket's TX path.
 *
 * @param xsk A pointer to the XSK socket info.
 * @param thread_id The thread ID to use to lookup the AF_XDP socket.
 * @param pckt The packet buffer starting at the Ethernet header.
 * @param length The packet buffer's length.
 * @param verbose Whether verbose is enabled or not.
 *
 * @return Returns 0 on success and -1 on failure.
 **/
int send_packet(xsk_socket_info_t *xsk, int thread_id, void *pckt, u16 length, u8 verbose)
{
    static __thread struct
    {
        void *packets[MAX_BATCH_SIZE];
        u16 lengths[MAX_BATCH_SIZE];
        u8 packet_data[MAX_BATCH_SIZE][MAX_PKT_LEN];
        int count;
    } batch = {0};

    // If packet is null and 0, try flushing.
    if (pckt == NULL && length == 0)
    {
        if (batch.count > 0)
        {
            int ret = send_packet_batch_internal(xsk, batch.packets, batch.lengths, batch.count);
            batch.count = 0;
            return ret;
        }
        return 0;
    }

    // Copy packet data to local buffer (since caller's buffer may be reused)
    memcpy(batch.packet_data[batch.count], pckt, length);
    batch.packets[batch.count] = batch.packet_data[batch.count];
    batch.lengths[batch.count] = length;
    batch.count++;

    // Send batch when full
    if (batch.count >= batch_size)
    {
        int ret = send_packet_batch_internal(xsk, batch.packets, batch.lengths, batch.count);
        batch.count = 0;

        return ret;
    }

    return 0;
}

/**
 * Sends multiple packets at once via batching.
 *
 * @param xsk A pointer to the XSK socket info.
 * @param pkts Array of packet buffers.
 * @param pkts_len Array of packet lengths.
 * @param amt Number of packets to send.
 *
 * @return Returns 0 on success and -1 on failure.
 **/
int send_packet_batch(xsk_socket_info_t *xsk, void *pkts, u16 *pkts_len, u16 amt)
{
    return send_packet_batch_internal(xsk, pkts, pkts_len, amt);
}

/**
 * Flushes any remaining packets in the send queue.
 *
 * @param xsk A pointer to the XSK socket info.
 *
 * @return Void
 **/
void flush_send_queue(xsk_socket_info_t *xsk)
{
    send_packet(xsk, 0, NULL, 0, 0);
}

/**
 * Retrieves the socket FD of XSK socket.
 *
 * @param xsk A pointer to the XSK socket info.
 *
 * @return The socket FD (-1 on failure)
 */
int get_socket_fd(xsk_socket_info_t *xsk)
{
    return xsk_socket__fd(xsk->xsk);
}

/**
 * Retrieves UMEM address at index we can fill with packet data.
 *
 * @param xsk A pointer to the XSK socket info.
 * @param idx The index we're retrieving (make sure it is below NUM_FRAMES).
 *
 * @return 64-bit address of location.
 **/
u64 get_umem_addr(xsk_socket_info_t *xsk, int idx)
{
    return xsk->umem_frame_addr[idx];
}

/**
 * Retrieves the memory location in the UMEM at address.
 *
 * @param xsk A pointer to the XSK socket info.
 * @param addr The address received by get_umem_addr.
 *
 * @return Pointer to address in memory of UMEM.
 **/
void *get_umem_loc(xsk_socket_info_t *xsk, u64 addr)
{
    return xsk_umem__get_data(xsk->umem->buffer, addr);
}

/**
 * Sets global variables from command line.
 *
 * @param cmd_af_xdp A pointer to the AF_XDP-specific command line variable.
 * @param verbose Whether we should print verbose.
 *
 * @return Void
 **/
void setup_af_xdp_variables(cli_af_xdp_t *cmd_af_xdp, int verbose)
{
    // Check for zero-copy or copy modes.
    if (cmd_af_xdp->zero_copy)
    {
        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets in zero-copy mode.\n");
        }

        bind_flags |= XDP_ZEROCOPY;
    }
    else if (cmd_af_xdp->copy)
    {
        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets in copy mode.\n");
        }

        bind_flags |= XDP_COPY;
    }

    // Check for no wakeup mode.
    if (cmd_af_xdp->no_wake_up)
    {
        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets in no wake-up mode.\n");
        }

        bind_flags &= ~XDP_USE_NEED_WAKEUP;
    }

    // Check for a static queue ID.
    if (cmd_af_xdp->queue_set)
    {
        static_queue_id = 1;
        queue_id = cmd_af_xdp->queue;

        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets with one queue ID => %d.\n", queue_id);
        }
    }

    // Check for shared UMEM.
    if (cmd_af_xdp->shared_umem)
    {
        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets with shared UMEM mode.\n");
        }

        is_shared_umem = 1;
        /* Note - Although documentation states to set bind flag XDP_SHARED_UMEM, this results in segfault and official sample with shared UMEMs does not do this. */
        // bind_flags |= XDP_SHARED_UMEM;
    }

    // Check for SKB mode.
    if (cmd_af_xdp->skb_mode)
    {
        if (verbose)
        {
            fprintf(stdout, "Running AF_XDP sockets in SKB mode.\n");
        }

        xdp_flags = XDP_FLAGS_SKB_MODE;
    }

    // Assign batch size.
    batch_size = cmd_af_xdp->batch_size;

    if (verbose)
    {
        fprintf(stdout, "Running AF_XDP sockets with batch size => %d.\n", batch_size);
    }
}

/**
 * Sets up UMEM at specific index.
 *
 * @param thread_id The thread ID/number.
 *
 * @return 0 on success and -1 on failure.
 **/
xsk_umem_info_t *setup_umem(int thread_id)
{
    // This indicates the buffer for frames and frame size for the UMEM area.
    void *frame_buffer;
    u64 frame_buffer_size = NUM_FRAMES * FRAME_SIZE;

    // Allocate blank memory space for the UMEM (aligned in chunks). Check as well.
    if (posix_memalign(&frame_buffer, getpagesize(), frame_buffer_size))
    {
        fprintf(stderr, "Could not allocate buffer memory for UMEM index #%d => %s (%d).\n", thread_id, strerror(errno), errno);

        return NULL;
    }

    return configure_xsk_umem(frame_buffer, frame_buffer_size);
}

/**
 * Sets up XSK (AF_XDP) socket.
 *
 * @param dev The interface the XDP program exists on (string).
 * @param thread_id The thread ID/number.
 * @param verbose Whether verbose mode is enabled.
 *
 * @return Returns the AF_XDP's socket FD or -1 on failure.
 **/
xsk_socket_info_t *setup_socket(const char *dev, u16 thread_id, int verbose)
{
    // Verbose message.
    if (verbose)
    {
        fprintf(stdout, "Attempting to setup AF_XDP socket. Dev => %s. Thread ID => %d.\n", dev, thread_id);
    }

    // Configure and create the AF_XDP/XSK socket.
    xsk_umem_info_t *umem;

    // Check for shared UMEM.
    if (is_shared_umem)
    {
        // Check if we need to allocate shared UMEM.
        if (shared_umem == NULL)
        {
            shared_umem = setup_umem(thread_id);

            if (shared_umem == NULL)
            {
                fprintf(stderr, "Failed to setup shared UMEM.\n");

                return NULL;
            }
        }

        umem = shared_umem;
    }
    else
    {
        // Otherwise, allocate our own UMEM for this thread/socket.
        umem = setup_umem(thread_id);
    }

    // Although this shouldn't happen, just check here in-case.
    if (umem == NULL)
    {
        fprintf(stderr, "UMEM at index 0 is NULL. Aborting...\n");

        return NULL;
    }

    xsk_socket_info_t *xsk = xsk_configure_socket(umem, (static_queue_id) ? queue_id : thread_id, (const char *)dev);

    // Check to make sure it's valid.
    if (xsk == NULL)
    {
        fprintf(stderr, "Could not setup AF_XDP socket at index %d :: %s (%d).\n", thread_id, strerror(errno), errno);

        return xsk;
    }

    // Retrieve the AF_XDP/XSK's socket FD and do a verbose print.
    int fd = xsk_socket__fd(xsk->xsk);

    if (verbose)
    {
        fprintf(stdout, "Created AF_XDP socket at index %d (FD => %d).\n", thread_id, fd);
    }

    // Return XSK socket.
    return xsk;
}

/**
 * Cleans up a specific AF_XDP/XSK socket.
 *
 * @param xsk A pointer to the XSK socket info.
 *
 * @return Void
 **/
void cleanup_socket(xsk_socket_info_t *xsk)
{
    // If the AF_XDP/XSK socket isn't NULL, delete it.
    if (xsk->xsk != NULL)
    {
        xsk_socket__delete(xsk->xsk);
    }

    // If the UMEM isn't NULL, delete it.
    if (xsk->umem != NULL)
    {
        xsk_umem__delete(xsk->umem->umem);
    }
}