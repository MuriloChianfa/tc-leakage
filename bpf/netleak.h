#ifndef __NETLEAK_H__
#define __NETLEAK_H__

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

/* Policy flags */
#define FLAG_KILL_SWITCH (1 << 0)

/* Per-cgroup routing policy */
struct policy {
    __u32 fwmark; /* mark applied to skb (non-zero) */
    __u32 flags;  /* bitmask: FLAG_KILL_SWITCH, etc. */
};

/*
 * Hash map: cgroup_id -> policy
 * Populated by userspace; read by the cgroup-skb egress program.
 */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u64);             /* cgroup v2 ID */
    __type(value, struct policy);
    __uint(max_entries, 256);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} cgroup_policy_map SEC(".maps");

#endif /* __NETLEAK_H__ */
