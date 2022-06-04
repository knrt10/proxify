<p align="center">
  <img src="https://user-images.githubusercontent.com/24803604/169649430-56ec4424-ff52-4559-933b-a6b105829861.png" />
</p>

> Proxy your raw TCP requests


# Contents

- [Demo](#demo)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Running the code](#running-the-code)
- [Internals](#internals)
- [eBPF](#ebpf)
    - [Testing](#testing)
- [TODO](#todo)
- [Inspiration](#inspiration)

## Demo
[![Demo](https://asciinema.org/a/499416.svg)](https://asciinema.org/a/499416?autoplay=1)

## Features

- `proxify` is a simple tool, which routes connection on specific port to given targets.
- It supports `eBPF steering` which routes connections on any configured port to proxifies's listener port.

## Prerequisites

You'll need a Linux host running kernel >= 5.9 (when they introduced sk_lookup) to build and run your BPF program. Linux in Docker on an M1 Mac will _not_ work. If you don't have access to a Linux box and can't run a VM you can spin up a dev VM.

You'll need the following tools and libraries installed:
  - `bpftool` compiled for a >= 5.9 kernel, because pre-5.9 `bpftool` doesn't know what an sk_lookup program is.
  - `libbpf` source code, which you can get from Github, because it has a recent `bpf_helper_defs.h` with `bpf_sk_assign` in it, which you need to make this program work.
  - clang>10 to generate ELF .o's that new bpftool will load from.

- [go-bindata](https://github.com/go-bindata/go-bindata/): Install using `go install -a -v github.com/go-bindata/go-bindata/...@latest`

## Usage

```
➜  ./proxify --help
Usage of ./proxify:
  -b	enable bpf steering
  -p int
    	listener port for bpf steering (default 8080)
```

## Running the code

```bash
# Build the proxify binary
make proxify

# To run in normal mode. Open up terminal and run the following command
./proxify

# To run in eBPF steering mode. Open up terminal and run the following.
sudo ./proxify -b

# On a different termial, run the following command
echo "hello there general kenobi" | nc -N -4 localhost 5001
```

## Internals

See [docs/architecture](docs/architecture.md)

## eBPF

### Testing

```bash
# Run in eBPF steering mode. Open up terminal and run the following.
sudo ./proxify -b

# On a different termial, run the following command
echo "hello there general kenobi" | nc -N -4 localhost 5001

# Check bpfmaps
sudo bpftool map

# Dump map data from id
sudo bptfool map dump id <id_number>

# Update new data to proxy_ports map. This adds port 7 to it.
sudo bpftool map update id <id_number> key 0x07 0x00 value 0x00

# Test the connection, it should work
echo "hello there general kenobi" | nc -N -4 localhost 7

# Check link
sudo bpftool link
```

## TODO
- [ ] Add tests
- [ ] Add github actions
- [ ] Update bpf maps on reload
- [ ] Add monitoring
- [ ] Add structured logging
- [ ] Background Health checks for unhealthy targets. 
- [ ] Filter requests from the start from unverified sources.
- [ ] Encryption/decryption support for requests.
- [ ] Caching support.
- [ ] Support batch request instead of just single request.
- [ ] Add security by adding authentication for client requests using certificates.

## Inspiration

[eBPF summit 2020](https://ebpf.io/summit-2020-slides/eBPF_Summit_2020-Lightning-Jakub_Sitnicki-Steering_connections_to_sockets_with_BPF_socke_lookup_hook.pdf)
