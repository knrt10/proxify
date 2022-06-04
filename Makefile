KERNEL := linux-5.13
KERNEL_INC := $(KERNEL)/usr/include
LIBBPF_INC := $(KERNEL)/tools/lib
LIBBPF_LIB := $(KERNL)/tools/lib/bpf/libbpf.a

CC := clang
CFLAGS := -g -O2 -Wall -Wextra
CPPFLAGS := -I$(KERNEL_INC) -I$(LIBBPF_INC)

proxify: pkg/assets/proxyassets-bpf.go
	go build -o proxify cmd/proxy/main.go

pkg/assets/proxyassets-bpf.go: bpf/bpf_proxy_dispatch.c $(KERNEL_INC) $(LIBBPF_INC) $(LIBBPF_LIB)
	$(CC) $(CPPFLAGS) $(CFLAGS) -target bpf -c bpf/bpf_proxy_dispatch.c -o bpf/bpf_proxy_dispatch.o
	go-bindata -pkg assets -prefix bpf -modtime 1 -o pkg/assets/proxyassets-bpf.go bpf/bpf_proxy_dispatch.o

clean:
	rm -f proxify pkg/assets/proxyassets-bpf.go

# Download kernel sources
$(KERNEL).tar.xz:
	curl -O https://cdn.kernel.org/pub/linux/kernel/v5.x/$(KERNEL).tar.xz

# Unpack kernel sources
$(KERNEL): $(KERNEL).tar.xz
	tar axf $<

# Install kernel headers
$(KERNEL_INC): $(KERNEL)
	make -C $< headers_install INSTALL_HDR_PATH=$@

# Build libbpf to generate helper definitions header
$(LIBBPF_LIB): $(KERNEL)
	make -C $</tools/lib/bpf
