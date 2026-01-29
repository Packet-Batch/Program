#pragma once

#include <helpers/int_types.h>

#include <constants.h>

#include <common/config.h>

#ifdef ENABLE_AF_XDP
#include <tech/af-xdp/cli.h>
#endif

typedef struct cmd_line
{
    const char *config;
    unsigned int list : 1;
    unsigned int verbose : 1;
    unsigned int help : 1;
    unsigned int cli : 1;

    /* Sequence options. */
    char *interface;
    unsigned int is_interface : 1;

    unsigned int block : 1;
    unsigned int is_block : 1;

    unsigned int track : 1;
    unsigned int is_track : 1;

    unsigned int l4_csum : 1;
    unsigned int is_l4_csum : 1;

    u64 max_pckts;
    unsigned int is_max_pckts : 1;

    u64 max_bytes;
    unsigned int is_max_bytes : 1;

    u64 pps;
    unsigned int is_pps : 1;

    u64 bps;
    unsigned int is_bps : 1;

    u64 time;
    unsigned int is_time : 1;

    u64 delay;
    unsigned int is_delay : 1;

    u16 threads;
    unsigned int is_threads : 1;

    char *src_mac;
    unsigned int is_src_mac : 1;

    char *dst_mac;
    unsigned int is_dst_mac : 1;

    u16 ttl_min;
    unsigned int is_ttl_min : 1;

    u16 ttl_max;
    unsigned int is_ttl_max : 1;

    u16 id_min;
    unsigned int is_id_min : 1;

    u16 id_max;
    unsigned int is_id_max : 1;

    char *src_ip;
    unsigned int is_src_ip : 1;

    char *dst_ip;
    unsigned int is_dst_ip : 1;

    char *protocol;
    unsigned int is_protocol : 1;

    u8 tos;
    unsigned int is_tos : 1;

    unsigned int l3_csum : 1;
    unsigned int is_l3_csum : 1;

    u16 udp_src_port;
    unsigned int is_udp_src_port : 1;

    u16 udp_dst_port;
    unsigned int is_udp_dst_port : 1;

    u16 tcp_src_port;
    unsigned int is_tcp_src_port : 1;

    u16 tcp_dst_port;
    unsigned int is_tcp_dst_port : 1;

    unsigned int tcp_syn : 1;
    unsigned int is_tcp_syn : 1;

    unsigned int tcp_ack : 1;
    unsigned int is_tcp_ack : 1;

    unsigned int tcp_psh : 1;
    unsigned int is_tcp_psh : 1;

    unsigned int tcp_rst : 1;
    unsigned int is_tcp_rst : 1;

    unsigned int tcp_fin : 1;
    unsigned int is_tcp_fin : 1;

    unsigned int tcp_urg : 1;
    unsigned int is_tcp_urg : 1;

    unsigned int tcp_ece : 1;
    unsigned int is_tcp_ece : 1;

    unsigned int tcp_cwr : 1;
    unsigned int is_tcp_cwr : 1;

    unsigned int tcp_cooked : 1;
    unsigned int is_tcp_cooked : 1;

    unsigned int tcp_one_connection : 1;
    unsigned int is_tcp_one_connection : 1;

    u8 icmp_code : 1;
    unsigned int is_icmp_code : 1;

    u8 icmp_type : 1;
    unsigned int is_icmp_type : 1;

    u16 pl_min_len;
    unsigned int is_pl_min_len : 1;

    u16 pl_max_len;
    unsigned int is_pl_max_len : 1;

    unsigned int pl_is_static : 1;
    unsigned int is_pl_is_static : 1;

    char *pl_exact;
    unsigned int is_pl_exact : 1;

    unsigned int pl_is_file : 1;
    unsigned int is_pl_is_file : 1;

    unsigned int pl_is_string : 1;
    unsigned int is_pl_is_string : 1;

#ifdef ENABLE_AF_XDP
    struct cli_af_xdp tech_af_xdp;
#endif
} cmd_line_t;

void print_cmd_help();
void parse_cli(int argc, char *argv[], struct cmd_line *cmd);
void parse_cli_seq_opts(struct cmd_line *cmd, struct config *cfg);