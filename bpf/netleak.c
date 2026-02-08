// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2024-2026 MuriloChianfa

#include "netleak.h"

/*
 * cgroup/sock_create program.
 *
 * Fires when any socket is created inside the cgroup. Sets sk->sk_mark
 * to the configured fwmark so that ALL subsequent route lookups for this
 * socket use the policy routing table. This is the hook that actually
 * steers traffic through the target interface.
 */
SEC("cgroup/sock_create")
int netleak_sock_create(struct bpf_sock *sk) {
  __u64 cgid = bpf_get_current_cgroup_id();

  struct policy *pol = bpf_map_lookup_elem(&cgroup_policy_map, &cgid);
  if (!pol)
    return 1; /* no policy -> allow */

  sk->mark = pol->fwmark;
  return 1; /* allow socket creation */
}

/*
 * cgroup_skb/egress program.
 *
 * Attached to the same cgroup. Enforces the kill-switch: when the target
 * interface is down, userspace sets FLAG_KILL_SWITCH in the map entry
 * and this program drops all egress packets.
 */
SEC("cgroup_skb/egress")
int netleak_egress(struct __sk_buff *skb) {
  (void)skb;
  __u64 cgid = bpf_get_current_cgroup_id();

  struct policy *pol = bpf_map_lookup_elem(&cgroup_policy_map, &cgid);
  if (!pol)
    return 1; /* no policy -> allow */

  if (pol->flags & FLAG_KILL_SWITCH)
    return 0; /* interface down -> drop */

  return 1; /* allow */
}

char _license[] SEC("license") = "GPL";
