ARCH_INCLUDES := /usr/include/$(shell dpkg-architecture -qDEB_HOST_MULTIARCH 2>/dev/null || echo x86_64-linux-gnu)

BPF_CFLAGS := -O2 -g \
  -target bpf \
  -Wall -Wextra \
  -I/usr/include/bpf \
  -I$(ARCH_INCLUDES)

BPF_SRCS := bpf/netleak.c
BPF_OBJS := bpf/netleak.o
TARGET   := netleak

.PHONY: all cmd clean

all: cmd $(BPF_OBJS)

%.o: %.c
	clang $(BPF_CFLAGS) -c $< -o $@

cmd:
	go build -C cmd -o ../$(TARGET)

clean:
	rm -f $(BPF_OBJS)
	rm -f $(TARGET)

show:
	sudo bpftool prog show
	sudo bpftool map list

logs:
	echo 1 | sudo tee /sys/kernel/debug/tracing/tracing_on
	sudo cat /sys/kernel/debug/tracing/trace_pipe
