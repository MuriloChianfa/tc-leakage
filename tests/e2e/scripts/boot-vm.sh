#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Boot a QEMU VM, provision it with Ansible, run a VPN test scenario, then tear down.
#
# Usage: boot-vm.sh <image> <seed.iso> <ssh-key> <ssh-port> <scenario>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${E2E_DIR}/../.." && pwd)"

IMAGE="$1"
SEED_ISO="$2"
SSH_KEY="$3"
SSH_PORT="$4"
SCENARIO="$5"

TIMEOUT="${E2E_TIMEOUT:-300}"
QEMU_MEM="${E2E_QEMU_MEM:-1024}"
QEMU_CPUS="${E2E_QEMU_CPUS:-2}"
QEMU_PID=""
WORK_DIR=""

# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

cleanup() {
    if [ -n "${QEMU_PID}" ] && kill -0 "${QEMU_PID}" 2>/dev/null; then
        echo "Shutting down QEMU (pid ${QEMU_PID})..."
        kill "${QEMU_PID}" 2>/dev/null || true
        wait "${QEMU_PID}" 2>/dev/null || true
    fi
    if [ -n "${WORK_DIR}" ] && [ -d "${WORK_DIR}" ]; then
        rm -rf "${WORK_DIR}"
    fi
}
trap cleanup EXIT

WORK_DIR=$(mktemp -d)
cp "${E2E_DIR}/${IMAGE}" "${WORK_DIR}/disk.qcow2"
qemu-img resize "${WORK_DIR}/disk.qcow2" 4G >/dev/null 2>&1 || true

KVM_FLAG=$(detect_kvm)

mkdir -p "${E2E_DIR}/logs"
QEMU_LOG="${E2E_DIR}/logs/${SCENARIO}-qemu.log"

# shellcheck disable=SC2086
qemu-system-x86_64 \
    ${KVM_FLAG} \
    -machine q35 \
    -m "${QEMU_MEM}" \
    -smp "${QEMU_CPUS}" \
    -nographic \
    -drive file="${WORK_DIR}/disk.qcow2",if=virtio,format=qcow2 \
    -drive file="${E2E_DIR}/${SEED_ISO}",if=virtio,media=cdrom \
    -nic user,model=virtio-net-pci,hostfwd=tcp::"${SSH_PORT}"-:22 \
    -serial mon:stdio \
    > "${QEMU_LOG}" 2>&1 &
QEMU_PID=$!

echo "QEMU started (pid ${QEMU_PID}), waiting for SSH..."

if ! wait_for_ssh "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" 120; then
    echo "VM failed to boot. QEMU log tail:"
    tail -30 "${QEMU_LOG}"
    exit 1
fi

echo "Building static netleak binary for VM..."
CGO_ENABLED=0 go build -C "${REPO_ROOT}/cmd" -buildvcs=false -ldflags '-s -w' -o "${REPO_ROOT}/netleak-static" 2>/dev/null

echo "Copying netleak binary and BPF object into VM..."
vm_scp "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" "${REPO_ROOT}/netleak-static" "/usr/local/bin/netleak"
vm_ssh "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" "mkdir -p /usr/lib/netleak"
vm_scp "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" "${REPO_ROOT}/bpf/netleak.o" "/usr/lib/netleak/netleak.o"
vm_ssh "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" "chmod +x /usr/local/bin/netleak"

echo "Running Ansible provisioning for scenario: ${SCENARIO}..."
ANSIBLE_CONFIG="${E2E_DIR}/ansible/ansible.cfg" \
    ansible-playbook \
    -i "${E2E_DIR}/ansible/inventory.yml" \
    --extra-vars "ssh_port=${SSH_PORT} ssh_key=${E2E_DIR}/${SSH_KEY}" \
    "${E2E_DIR}/ansible/playbooks/provision-base.yml"

ANSIBLE_CONFIG="${E2E_DIR}/ansible/ansible.cfg" \
    ansible-playbook \
    -i "${E2E_DIR}/ansible/inventory.yml" \
    --extra-vars "ssh_port=${SSH_PORT} ssh_key=${E2E_DIR}/${SSH_KEY}" \
    "${E2E_DIR}/ansible/playbooks/provision-${SCENARIO}.yml"

echo "Running test assertions for scenario: ${SCENARIO}..."
TEST_LOG="${E2E_DIR}/logs/${SCENARIO}-test.log"

if timeout "${TIMEOUT}" bash "${SCRIPT_DIR}/test-${SCENARIO}.sh" \
    "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" 2>&1 | tee "${TEST_LOG}"; then
    echo "Scenario ${SCENARIO}: PASSED"
    exit 0
else
    echo "Scenario ${SCENARIO}: FAILED"
    echo "Collecting VM diagnostics..."
    vm_ssh "${E2E_DIR}/${SSH_KEY}" "${SSH_PORT}" \
        "ip addr; ip route; ip rule; dmesg | tail -50" \
        > "${E2E_DIR}/logs/${SCENARIO}-diagnostics.log" 2>&1 || true
    exit 1
fi
