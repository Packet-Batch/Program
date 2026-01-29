#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <linux/types.h>

#include <json-c/json.h>

#include <common/utils.h>

#include <common/config.h>

/**
 * Parses a config file including the main config options and sequences. It then fills out the config structure passed in the function's parameters.
 *
 * @param file_name The YAML config file to parse.
 * @param cfg A pointer to a config structure that'll be filled in with values.
 * @param only_seq If set to 1, this function will only parse sequences and add onto the number.
 * @param seq_num A pointer to the current sequence # (starting from 0).
 *
 * @return Returns 0 on success and -1 on failure.
 **/
int parse_config(const char file_name[], config_t *cfg, int only_seq, int *seq_num, u8 log)
{
    // Attempt to open config file with JSON.
    json_object *root = json_object_from_file(file_name);

    if (root == NULL)
    {
        fprintf(stderr, "Failed to open config file '%s'.\n", file_name);

        return 1;
    }

    json_object *tmp_obj;

    // Parse global interface.
    if (json_object_object_get_ex(root, "interface", &tmp_obj))
    {
        cfg->interface = (char *)json_object_get_string(tmp_obj);
    }

    // Parse sequences array.
    json_object *j_sequences;

    if (!json_object_object_get_ex(root, "sequences", &j_sequences))
    {
        fprintf(stderr, "Failed to open sequences array.\n");

        return 1;
    }

    // Check length of sequences.
    int seq_len = json_object_array_length(j_sequences);

    if (seq_len > 0)
    {
        // Loop through each sequence.
        for (int i = 0; i < seq_len; i++)
        {
            // Retrieve current sequence.
            sequence_t *seq = &cfg->seq[i];

            // Get sequence JSON object.
            json_object *seq_obj = json_object_array_get_idx(j_sequences, i);

            // Retrieve interface.
            if (json_object_object_get_ex(seq_obj, "interface", &tmp_obj))
            {
                seq->interface = (char *)json_object_get_string(tmp_obj);
            }

            // Retrieve block.
            if (json_object_object_get_ex(seq_obj, "block", &tmp_obj))
            {
                seq->block = json_object_get_boolean(tmp_obj);
            }

            // Retrieve max packets.
            if (json_object_object_get_ex(seq_obj, "maxpckts", &tmp_obj))
            {
                seq->max_pckts = json_object_get_uint64(tmp_obj);
            }

            // Retrieve max bytes.
            if (json_object_object_get_ex(seq_obj, "maxbytes", &tmp_obj))
            {
                seq->max_bytes = json_object_get_uint64(tmp_obj);
            }

            // Retrieve packets per second rate.
            if (json_object_object_get_ex(seq_obj, "pps", &tmp_obj))
            {
                seq->pps = json_object_get_uint64(tmp_obj);
            }

            // Retrieve bytes per second rate.
            if (json_object_object_get_ex(seq_obj, "bps", &tmp_obj))
            {
                seq->bps = json_object_get_uint64(tmp_obj);
            }

            // Retrieve time
            if (json_object_object_get_ex(seq_obj, "time", &tmp_obj))
            {
                seq->time = json_object_get_int(tmp_obj);
            }

            // Retrieve threads.
            if (json_object_object_get_ex(seq_obj, "threads", &tmp_obj))
            {
                seq->threads = json_object_get_int(tmp_obj);
            }

            // Retrieve delay.
            if (json_object_object_get_ex(seq_obj, "delay", &tmp_obj))
            {
                seq->delay = json_object_get_uint64(tmp_obj);
            }

            // Retrieve tracking.
            if (json_object_object_get_ex(seq_obj, "track", &tmp_obj))
            {
                seq->track = json_object_get_boolean(tmp_obj);
            }

            // Retrieve layer-4 checksum.
            if (json_object_object_get_ex(seq_obj, "l4csum", &tmp_obj))
            {
                seq->l4_csum = json_object_get_boolean(tmp_obj);
            }

            // Retrieve ethernet object.
            json_object *eth_obj;

            if (json_object_object_get_ex(seq_obj, "eth", &eth_obj))
            {
                // Source MAC address.
                if (json_object_object_get_ex(eth_obj, "smac", &tmp_obj))
                {
                    seq->eth.src_mac = (char *)json_object_get_string(tmp_obj);
                }

                // Destination MAC address.
                if (json_object_object_get_ex(eth_obj, "dmac", &tmp_obj))
                {
                    seq->eth.dst_mac = (char *)json_object_get_string(tmp_obj);
                }
            }

            // Retrieve IP object.
            json_object *ip_obj;

            if (json_object_object_get_ex(seq_obj, "ip", &ip_obj))
            {
                // Source IP.
                if (json_object_object_get_ex(ip_obj, "sip", &tmp_obj))
                {
                    seq->ip.src_ip = (char *)json_object_get_string(tmp_obj);
                }

                // Destination IP.
                if (json_object_object_get_ex(ip_obj, "dip", &tmp_obj))
                {
                    seq->ip.dst_ip = (char *)json_object_get_string(tmp_obj);
                }

                // Protocol.
                if (json_object_object_get_ex(ip_obj, "protocol", &tmp_obj))
                {
                    seq->ip.protocol = (char *)json_object_get_string(tmp_obj);
                }

                // ToS.
                if (json_object_object_get_ex(ip_obj, "tos", &tmp_obj))
                {
                    seq->ip.tos = json_object_get_int(tmp_obj);
                }

                // Checksum.
                if (json_object_object_get_ex(ip_obj, "csum", &tmp_obj))
                {
                    seq->ip.csum = json_object_get_boolean(tmp_obj);
                }

                // TTL object.
                json_object *ttl_obj;

                if (json_object_object_get_ex(ip_obj, "ttl", &ttl_obj))
                {
                    // Minimum TTL.
                    if (json_object_object_get_ex(ttl_obj, "min", &tmp_obj))
                    {
                        seq->ip.min_ttl = json_object_get_int(tmp_obj);
                    }

                    // Maximum TTL.
                    if (json_object_object_get_ex(ttl_obj, "max", &tmp_obj))
                    {
                        seq->ip.max_ttl = json_object_get_int(tmp_obj);
                    }
                }

                // ID object.
                json_object *id_obj;

                if (json_object_object_get_ex(ip_obj, "id", &id_obj))
                {
                    // Minimum ID.
                    if (json_object_object_get_ex(id_obj, "min", &tmp_obj))
                    {
                        seq->ip.min_id = json_object_get_int(tmp_obj);
                    }

                    // Maximum ID.
                    if (json_object_object_get_ex(id_obj, "max", &tmp_obj))
                    {
                        seq->ip.max_id = json_object_get_int(tmp_obj);
                    }
                }

                // Ranges array.
                json_object *ranges_obj;

                if (json_object_object_get_ex(ip_obj, "ranges", &ranges_obj))
                {
                    int ranges_len = json_object_array_length(ranges_obj);

                    if (ranges_len > 0)
                    {
                        for (int j = 0; j < ranges_len; j++)
                        {
                            // Retrieve specific range and add to ranges array.
                            json_object *range_obj = json_object_array_get_idx(ranges_obj, j);

                            seq->ip.ranges[j] = (char *)json_object_get_string(range_obj);
                            seq->ip.range_count++;
                        }
                    }
                }
            }

            // UDP object.
            json_object *udp_obj;

            if (json_object_object_get_ex(seq_obj, "udp", &udp_obj))
            {
                // Source port.
                if (json_object_object_get_ex(udp_obj, "sport", &tmp_obj))
                {
                    seq->udp.src_port = json_object_get_int(tmp_obj);
                }

                // Destination port.
                if (json_object_object_get_ex(udp_obj, "dport", &tmp_obj))
                {
                    seq->udp.dst_port = json_object_get_int(tmp_obj);
                }
            }

            // TCP object.
            json_object *tcp_obj;

            if (json_object_object_get_ex(seq_obj, "tcp", &tcp_obj))
            {
                // Cooked socket.
                if (json_object_object_get_ex(tcp_obj, "cooked", &tmp_obj))
                {
                    seq->tcp.cooked = json_object_get_boolean(tmp_obj);
                }

                // One connection.
                if (json_object_object_get_ex(tcp_obj, "oneconn", &tmp_obj))
                {
                    seq->tcp.one_connection = json_object_get_boolean(tmp_obj);
                }

                // Source port.
                if (json_object_object_get_ex(tcp_obj, "sport", &tmp_obj))
                {
                    seq->tcp.src_port = json_object_get_int(tmp_obj);
                }

                // Destination port.
                if (json_object_object_get_ex(tcp_obj, "dport", &tmp_obj))
                {
                    seq->tcp.dst_port = json_object_get_int(tmp_obj);
                }

                // SYN flag.
                if (json_object_object_get_ex(tcp_obj, "syn", &tmp_obj))
                {
                    seq->tcp.syn = json_object_get_boolean(tmp_obj);
                }

                // PSH flag.
                if (json_object_object_get_ex(tcp_obj, "psh", &tmp_obj))
                {
                    seq->tcp.psh = json_object_get_boolean(tmp_obj);
                }

                // FIN flag.
                if (json_object_object_get_ex(tcp_obj, "fin", &tmp_obj))
                {
                    seq->tcp.fin = json_object_get_boolean(tmp_obj);
                }

                // ACK flag.
                if (json_object_object_get_ex(tcp_obj, "ack", &tmp_obj))
                {
                    seq->tcp.ack = json_object_get_boolean(tmp_obj);
                }

                // RST flag.
                if (json_object_object_get_ex(tcp_obj, "rst", &tmp_obj))
                {
                    seq->tcp.rst = json_object_get_boolean(tmp_obj);
                }

                // URG flag.
                if (json_object_object_get_ex(tcp_obj, "urg", &tmp_obj))
                {
                    seq->tcp.urg = json_object_get_boolean(tmp_obj);
                }

                // ECE flag.
                if (json_object_object_get_ex(tcp_obj, "ece", &tmp_obj))
                {
                    seq->tcp.ece = json_object_get_boolean(tmp_obj);
                }

                // CWR flag.
                if (json_object_object_get_ex(tcp_obj, "cwr", &tmp_obj))
                {
                    seq->tcp.cwr = json_object_get_boolean(tmp_obj);
                }
            }

            // ICMP object.
            json_object *icmp_obj;

            if (json_object_object_get_ex(seq_obj, "icmp", &icmp_obj))
            {
                // Code.
                if (json_object_object_get_ex(icmp_obj, "code", &tmp_obj))
                {
                    seq->icmp.code = json_object_get_int(tmp_obj);
                }

                // Type.
                if (json_object_object_get_ex(icmp_obj, "type", &tmp_obj))
                {
                    seq->icmp.type = json_object_get_int(tmp_obj);
                }
            }

            // Payloads object.
            json_object *pls_obj;

            if (json_object_object_get_ex(seq_obj, "payloads", &pls_obj))
            {
                int pls_len = json_object_array_length(pls_obj);

                if (pls_len > 0)
                {
                    for (int j = 0; j < pls_len; j++)
                    {
                        json_object *pl_obj = json_object_array_get_idx(pls_obj, j);

                        payload_opt_t *pl = &seq->pls[j];

                        // Is static.
                        if (json_object_object_get_ex(pl_obj, "isstatic", &tmp_obj))
                        {
                            pl->is_static = json_object_get_boolean(tmp_obj);
                        }

                        // Is file.
                        if (json_object_object_get_ex(pl_obj, "isfile", &tmp_obj))
                        {
                            pl->is_file = json_object_get_boolean(tmp_obj);
                        }

                        // Is string.
                        if (json_object_object_get_ex(pl_obj, "isstring", &tmp_obj))
                        {
                            pl->is_string = json_object_get_boolean(tmp_obj);
                        }

                        // Exact.
                        if (json_object_object_get_ex(pl_obj, "exact", &tmp_obj))
                        {
                            pl->exact = (char *)json_object_get_string(tmp_obj);
                        }

                        // Length object.
                        json_object *len_obj;

                        if (json_object_object_get_ex(pl_obj, "length", &len_obj))
                        {
                            // Minimum length.
                            if (json_object_object_get_ex(len_obj, "min", &tmp_obj))
                            {
                                pl->min_len = json_object_get_int(tmp_obj);
                            }

                            // Maximum length.
                            if (json_object_object_get_ex(len_obj, "max", &tmp_obj))
                            {
                                pl->max_len = json_object_get_int(tmp_obj);
                            }
                        }

                        seq->pl_cnt++;
                    }
                }
            }

            *seq_num += 1;
        }
    }

    return 0;
}

/**
 * Clears a sequence.
 *
 * @param cfg A pointer to the config structure.
 * @param seq_num Which sequence to reset.
 *
 * @return Void
 **/
void clear_sequence(config_t *cfg, int seq_num)
{
    sequence_t *seq = &cfg->seq[seq_num];

    seq->interface = NULL;
    seq->block = 1;
    seq->track = 0;
    seq->max_pckts = 0;
    seq->max_bytes = 0;
    seq->pps = 0;
    seq->bps = 0;
    seq->threads = 0;
    seq->time = 0;
    seq->delay = 1000000;

    seq->eth.src_mac = NULL;
    seq->eth.dst_mac = NULL;

    seq->ip.src_ip = NULL;
    seq->ip.dst_ip = NULL;
    seq->ip.protocol = NULL;
    seq->ip.tos = 0;
    seq->ip.min_ttl = 64;
    seq->ip.max_ttl = 64;
    seq->ip.min_id = 0;
    seq->ip.max_id = 64000;
    seq->ip.csum = 1;

    seq->udp.src_port = 0;
    seq->udp.dst_port = 0;

    seq->tcp.syn = 0;
    seq->tcp.ack = 0;
    seq->tcp.psh = 0;
    seq->tcp.rst = 0;
    seq->tcp.fin = 0;
    seq->tcp.urg = 0;
    seq->tcp.cooked = 0;
    seq->tcp.one_connection = 0;

    seq->icmp.code = 0;
    seq->icmp.type = 0;

    seq->l4_csum = 1;

    // Reset includes.
    for (int i = 0; i < MAX_INCLUDES; i++)
    {
        seq->includes[i] = NULL;
    }

    // Reset IP ranges.
    for (int i = 0; i < MAX_RANGES; i++)
    {
        seq->ip.ranges[i] = NULL;
    }

    // Reset payloads.
    for (int i = 0; i < MAX_PAYLOADS; i++)
    {
        payload_opt_t *pl = &seq->pls[i];

        pl->exact = NULL;
        pl->is_file = 0;
        pl->is_static = 0;
        pl->is_static = 0;
        pl->min_len = 0;
        pl->max_len = 0;
    }
}

/**
 * Clears a sequence.
 *
 * @param cfg A pointer to the config structure.
 * @param seq_cnt How many sequences we have.
 *
 * @return Void
 **/
void print_config(config_t *cfg, int seq_cnt)
{
    fprintf(stdout, "Found %d sequences.\n", seq_cnt);

    fprintf(stdout, "Got interface => %s.\n", cfg->interface ? cfg->interface : "N/A");

    fprintf(stdout, "Sequences:\n\n--------------------------\n");

    for (int i = 0; i < seq_cnt; i++)
    {
        sequence_t *seq = &cfg->seq[i];

        if (!seq)
            continue;

        fprintf(stdout, "Sequence #%d:\n", i);

        // General settings.
        fprintf(stdout, "\tGeneral\n");
        fprintf(stdout, "\t\tIncludes =>\n");

        if (seq->include_count > 0)
        {
            for (int j = 0; j < seq->include_count; j++)
            {
                fprintf(stdout, "\t\t\t- %s\n", seq->includes[j]);
            }
        }

        fprintf(stdout, "\t\tInterface Override => %s\n", seq->interface ? seq->interface : "N/A");
        fprintf(stdout, "\t\tBlock => %s\n", seq->block ? "Yes" : "No");
        fprintf(stdout, "\t\tTrack => %s\n", seq->track ? "Yes" : "No");
        fprintf(stdout, "\t\tMax Packets => %llu\n", seq->max_pckts);
        fprintf(stdout, "\t\tMax Bytes => %llu.\n", seq->max_bytes);
        fprintf(stdout, "\t\tPackets Per Second => %llu.\n", seq->pps);
        fprintf(stdout, "\t\tBytes Per Second => %llu\n", seq->bps);
        fprintf(stdout, "\t\tTime => %llu\n", seq->time);
        fprintf(stdout, "\t\tDelay => %llu\n", seq->delay);
        fprintf(stdout, "\t\tThreads => %u\n", seq->threads);

        // Ethernet settings.
        fprintf(stdout, "\t\tEthernet\n");
        fprintf(stdout, "\t\t\tSource MAC => %s\n", seq->eth.src_mac ? seq->eth.src_mac : "N/A");
        fprintf(stdout, "\t\t\tDestination MAC => %s\n", seq->eth.dst_mac ? seq->eth.dst_mac : "N/A");

        // IP settings.
        fprintf(stdout, "\t\tIP\n");
        fprintf(stdout, "\t\t\tProtocol => %s\n", seq->ip.protocol ? seq->ip.protocol : "N/A");

        fprintf(stdout, "\t\t\tSource IP => %s\n", seq->ip.src_ip ? seq->ip.src_ip : "N/A");
        fprintf(stdout, "\t\t\tDestination IP => %s\n", seq->ip.dst_ip ? seq->ip.dst_ip : "N/A");

        fprintf(stdout, "\t\t\tType of Service => %d\n", seq->ip.tos);
        fprintf(stdout, "\t\t\tMin TTL => %d\n", seq->ip.min_ttl);
        fprintf(stdout, "\t\t\tMax TTL => %d\n", seq->ip.max_ttl);
        fprintf(stdout, "\t\t\tMin ID => %d\n", seq->ip.min_id);
        fprintf(stdout, "\t\t\tMax ID => %d\n", seq->ip.max_id);
        fprintf(stdout, "\t\t\tChecksum => %s\n", seq->ip.csum ? "Yes" : "No");

        if (seq->ip.range_count > 0)
        {
            fprintf(stdout, "\t\t\tRanges:\n");

            for (int j = 0; j < seq->ip.range_count; j++)
            {
                fprintf(stdout, "\t\t\t\t- %s\n", seq->ip.ranges[j]);
            }
        }

        // TCP settings.
        fprintf(stdout, "\t\tTCP\n");
        fprintf(stdout, "\t\t\tSource Port => %d\n", seq->tcp.src_port);
        fprintf(stdout, "\t\t\tDest Port => %d\n", seq->tcp.dst_port);
        fprintf(stdout, "\t\t\tUse Socket => %s\n", seq->tcp.cooked ? "Yes" : "No");
        fprintf(stdout, "\t\t\tOne Connection => %s\n", seq->tcp.one_connection ? "Yes" : "No");
        fprintf(stdout, "\t\t\tSYN Flag => %s\n", seq->tcp.syn ? "Yes" : "No");
        fprintf(stdout, "\t\t\tPSH Flag => %s\n", seq->tcp.psh ? "Yes" : "No");
        fprintf(stdout, "\t\t\tFIN Flag => %s\n", seq->tcp.fin ? "Yes" : "No");
        fprintf(stdout, "\t\t\tACK Flag => %s\n", seq->tcp.ack ? "Yes" : "No");
        fprintf(stdout, "\t\t\tRST Flag => %s\n", seq->tcp.rst ? "Yes" : "No");
        fprintf(stdout, "\t\t\tURG Flag => %s\n", seq->tcp.urg ? "Yes" : "No");
        fprintf(stdout, "\t\t\tECE Flag => %s\n", seq->tcp.ece ? "Yes" : "No");
        fprintf(stdout, "\t\t\tCWR Flag => %s\n", seq->tcp.cwr ? "Yes" : "No");

        // UDP settings.
        fprintf(stdout, "\t\tUDP\n");
        fprintf(stdout, "\t\t\tSrc Port => %d\n", seq->udp.src_port);
        fprintf(stdout, "\t\t\tDst Port => %d\n", seq->udp.dst_port);

        // ICMP settings.
        fprintf(stdout, "\t\tICMP\n");
        fprintf(stdout, "\t\t\tCode => %d\n", seq->icmp.code);
        fprintf(stdout, "\t\t\tType => %d\n", seq->icmp.type);

        // Layer 4 setting(s).
        fprintf(stdout, "\t\tLayer 4\n");
        fprintf(stdout, "\t\t\tChecksum => %s\n", seq->l4_csum ? "Yes" : "No");

        if (seq->pl_cnt > 0)
        {
            fprintf(stdout, "\t\tPayloads (%d)\n", seq->pl_cnt);

            for (int j = 0; j < seq->pl_cnt; j++)
            {
                payload_opt_t *pl = &seq->pls[j];

                fprintf(stdout, "\t\t\t#%d\n", j + 1);

                fprintf(stdout, "\t\t\t\tMin Length => %d\n", pl->min_len);
                fprintf(stdout, "\t\t\t\tMax Length => %d\n", pl->max_len);
                fprintf(stdout, "\t\t\t\tIs Static => %s\n", pl->is_static ? "Yes" : "No");
                fprintf(stdout, "\t\t\t\tIs File => %s\n", pl->is_file ? "Yes" : "No");
                fprintf(stdout, "\t\t\t\tIs String => %s\n", pl->is_string ? "Yes" : "No");
                fprintf(stdout, "\t\t\t\tExact String => %s\n", pl->exact ? pl->exact : "N/A");
            }
        }

        fprintf(stdout, "\n\n");
    }
}