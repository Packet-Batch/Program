#pragma once

// The max packet length. This shouldn't be changed unless you know what you're doing.
#define MAX_PCKT_LEN 0xFFFF

// The maximum amount of threads to support for sequences. 4096 should be enough I think... (lol)
#define MAX_THREADS 4096

// The maximum length of a name (e.g. device name).
#define MAX_NAME_LEN 64

// Enables AF_XDP socket support (required for now since we have no other techs \o/).
#define ENABLE_AF_XDP

// WARNING: Experimental
// When defined, the AF_XDP socket's UMEM frame buffers doesn't get the packet buffer copied to it after the first time. This should technically result in better throughput due to reduced memory copying, but only would work for completely static packets.
// #define AF_XDP_NO_REFILL_FRAMES

// The default config path to parse.
#define CONF_PATH_DEFAULT "/etc/pcktbatch/conf.json"

// If defined, the program will attempt to raise the stack limit. This shouldn't be needed, but wanted to include just in case.
// #define CONF_UNLOCK_RLIMIT

// If defined, the program uses a much more randomized seed by retrieving random bytes from `/dev/urandom`.
// This results in more randomness with each packet, but adds a bit more overhead.
// #define VERY_RANDOM

// If defined, if we only detect a single static packet being sent, we skip a lot of unnecessary steps.
#define SEQUENCE_USE_STATIC_PKT

// If defined, attempts to generate the most amount of packets per second by sending single empty packets (of any of the protocols) with static packets.
// All checks are also skipped including limits such as time, max packets, packet rates, etc. This is to achieve max throughput possible.
// #define SEQUENCE_MAX_PERFORMANCE