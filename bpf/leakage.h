#ifndef __LEAKAGE__

/* eBPF includes */
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf_tracing.h>
#include <bpf_endian.h>

/* Defining TC Actions */
/* https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/uapi/linux/pkt_cls.h */
#define TC_ACT_UNSPEC	(-1)
#define TC_ACT_OK		0
#define TC_ACT_RECLASSIFY	1
#define TC_ACT_SHOT		2
#define TC_ACT_PIPE		3
#define TC_ACT_STOLEN		4
#define TC_ACT_QUEUED		5
#define TC_ACT_REPEAT		6
#define TC_ACT_REDIRECT		7
#define TC_ACT_TRAP		8 
#define TC_ACT_VALUE_MAX	TC_ACT_TRAP

/* Hash map for PID to iFIndex mapping */
/* https://docs.kernel.org/accounting/taskstats.html */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __type(key, __u32);   // PID
    __type(value, __u32); // iFIndex
    __uint(max_entries, 1024);
} PID_IF_MAP SEC(".maps");

/* Hash map for TGID to iFIndex mapping: */
/* https://docs.kernel.org/accounting/taskstats.html */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __type(key, __u32);   // TGID
    __type(value, __u32); // iFIndex
    __uint(max_entries, 1024);
} TGID_IF_MAP SEC(".maps");

/* Parent Proccess struct */
struct trace_event_raw_sched_process_exit {
    __s64 state;
    __u32 pid;
    __u32 ppid;
    char comm[16];
};
#endif
