#include <leakage.h>

/* Tracepoint for process fork */
SEC("tracepoint/sched/sched_process_fork")
int handle_sched_process_fork(struct trace_event_raw_sched_process_fork *ctx) {
    __u32 parent_tgid = ctx->parent_pid; // TGID of the parent
    __u32 child_pid = ctx->child_pid;    // PID (TID) of the child

    // Lookup parent TGID in PID_IF_MAP
    __u32 *ifindex = bpf_map_lookup_elem(&PID_IF_MAP, &parent_tgid);
    if (ifindex) {
        // Map child PID's TGID (same as parent) to the same ifindex
        bpf_map_update_elem(&PID_IF_MAP, &parent_tgid, ifindex, BPF_ANY);
        bpf_printk("Mapped child PID %u to ifindex %u (parent TGID %u)\n", child_pid, *ifindex, parent_tgid);
    }

    return 0;
}

/* Tracepoint for process exit */
SEC("tracepoint/sched/sched_process_exit")
int handle_sched_process_exit(struct trace_event_raw_sched_process_exit *ctx) {
    __u32 exiting_pid = ctx->pid;  // TID of the exiting process
    __u32 exiting_tgid = ctx->ppid; // TGID of the exiting process (usually the parent)

    bpf_map_delete_elem(&PID_IF_MAP, &exiting_tgid);
    bpf_printk("Removed TGID %u from PID_IF_MAP on exit\n", exiting_tgid);

    return 0;
}

// Fallback to bpf_probe_read_kernel if CO-RE is problematic
static __always_inline __u64 my_get_current_pid_tgid(void) {
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    if (!task) {
        // bpf_printk("Task struct is NULL\n");
        return 0;
    }

    __u32 tgid = 0, pid = 0;

    if (bpf_probe_read_kernel(&tgid, sizeof(tgid), &task->tgid) < 0) {
        // bpf_printk("Failed to read TGID\n");
        return 0;
    }
    if (bpf_probe_read_kernel(&pid, sizeof(pid), &task->pid) < 0) {
        // bpf_printk("Failed to read PID\n");
        return 0;
    }

    // bpf_printk("TGID=%u, PID=%u\n", tgid, pid);
    return ((__u64)tgid << 32) | pid;
}

/*
 * TC Egress program to redirect packets based on PID -> ifindex mapping.
 */
SEC("tc_egress")
int tc_leakage() {
    __u64 pid_tgid = my_get_current_pid_tgid();
    __u64 pid = pid_tgid & 0xFFFFFFFF;

    __u32 *ifindex = bpf_map_lookup_elem(&PID_IF_MAP, &pid);
    if (ifindex) {
        bpf_printk("TC Redirect: Found ifindex %u for PID %u\n", *ifindex, pid);
        // Redirect to the specified interface
        int ret = bpf_redirect(*ifindex, 0);
        if (ret < 0) {
            bpf_printk("bpf_redirect failed with error %d\n", ret);
            return TC_ACT_SHOT;
        }
        return ret;
    }

    __u32 *ifindex_tgid = bpf_map_lookup_elem(&PID_IF_MAP, &pid_tgid);
    if (ifindex_tgid) {
        bpf_printk("TC Redirect: Found ifindex_tgid %u for TGID %u\n", *ifindex_tgid, pid);
        // Redirect to the specified interface
        int ret = bpf_redirect(*ifindex_tgid, 0);
        if (ret < 0) {
            bpf_printk("bpf_redirect failed with error %d\n", ret);
            return TC_ACT_SHOT;
        }
        return ret;
    }

    bpf_printk("TC Pass: No mapping found for PID or TGID %u\n", pid);
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
