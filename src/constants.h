#pragma once

#define ENABLE_AF_XDP

// WARNING: Experimental
// When defined, the AF_XDP socket's UMEM frame buffers doesn't get the packet buffer copied to it after the first time. This should technically result in better throughput due to reduced memory copying, but only would work for completely static packets.
// #define AF_XDP_NO_REFILL_FRAMES

#define CONF_PATH_DEFAULT "/etc/pcktbatch/conf.json"

// #define CONF_UNLOCK_RLIMIT