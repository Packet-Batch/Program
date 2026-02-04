#pragma once

#include <helpers/int_types.h>

#define MAX_INCLUDES 64
#define MAX_SEQUENCES 256
#define MAX_RANGES 64

#define MAX_PAYLOADS 256

typedef struct eth_opt
{
    char *src_mac;
    char *dst_mac;
} eth_opt_t;

typedef struct ip_opt
{
    // Protocol (Required).
    char *protocol;

    // Source and destination addresses (Required).
    char *src_ip;
    char *dst_ip;
    char *ranges[MAX_RANGES];
    u16 range_count;

    // Type of Service.
    u8 tos;

    // Time to Live.
    u8 min_ttl;
    u8 max_ttl;

    // ID.
    u16 min_id;
    u16 max_id;

    // Do checksum.
    unsigned int csum : 1;
} ip_opt_t;

typedef struct tcp_opt
{
    u16 src_port;
    u16 dst_port;

    // TCP flags.
    unsigned int syn : 1;
    unsigned int psh : 1;
    unsigned int fin : 1;
    unsigned int ack : 1;
    unsigned int rst : 1;
    unsigned int urg : 1;
    unsigned int ece : 1;
    unsigned int cwr : 1;

    unsigned int cooked : 1;
    unsigned int one_connection : 1;
} tcp_opt_t;

typedef struct udp_opt
{
    u16 src_port;
    u16 dst_port;
} udp_opt_t;

typedef struct icmp_opt
{
    u8 code;
    u8 type;
} icmp_opt_t;

typedef struct payload_opt
{
    u16 min_len;
    u16 max_len;
    unsigned int is_static : 1;
    unsigned int is_file : 1;
    unsigned int is_string : 1;
    char *exact;
} payload_opt_t;

typedef struct sequence
{
    // General options.
    char *interface;

    unsigned int block : 1;
    unsigned int track : 1;

    u64 max_pckts;
    u64 max_bytes;

    u64 pps;
    u64 bps;

    u64 time;
    u64 delay;
    u16 threads;
    char *includes[MAX_INCLUDES];
    u16 include_count;

    // Ethernet options.
    eth_opt_t eth;

    // IP options.
    ip_opt_t ip;

    // Layer 4 options.
    tcp_opt_t tcp;
    udp_opt_t udp;
    icmp_opt_t icmp;
    unsigned int l4_csum : 1;

    // Payload options.
    int pl_cnt;
    payload_opt_t pls[MAX_PAYLOADS];
} sequence_t;

typedef struct config
{
    // Device options.
    char *interface;

    sequence_t seq[MAX_SEQUENCES];
} config_t;

int config__parse(const char file_name[], struct config *cfg, int config__only_seq, int *seq_num, u8 log);
void config__clr_seq(struct config *cfg, int seq_num);
void config__print(struct config *cfg, int seq_cnt);