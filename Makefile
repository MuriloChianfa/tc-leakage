# bpf/Makefile

BPF_CFLAGS := -O2 -g \
  -target bpf \
  -Wextra \
  -I./bpf/ -I/usr/include/bpf

BPF_SRCS := bpf/leakage.c
BPF_OBJS := bpf/leakage.o
BPF_IF   := enp2s0
TARGET   := tc-leakage

.PHONY: cmd

all: cmd $(BPF_OBJS)

%.o: %.c
	clang $(BPF_CFLAGS) -c $< -o $@

cmd:
	go build -C cmd -o ../$(TARGET)

load:
	sudo tc qdisc add dev $(BPF_IF) clsact
	sudo tc filter add dev $(BPF_IF) egress bpf da obj bpf/leakage.o sec tc_egress

unload:
	sudo tc qdisc delete dev $(BPF_IF) clsact 2>/dev/null
	sudo tc filter delete dev $(BPF_IF) egress bpf 2>/dev/null

clean:
	rm -f $(BPF_OBJS)
	rm -f $(TARGET)

show:
	sudo bpftool prog show
	sudo bpftool map list

logs:
	echo 1 | sudo tee /sys/kernel/debug/tracing/tracing_on
	sudo cat /sys/kernel/debug/tracing/trace_pipe
