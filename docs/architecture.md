## Go

- When app is started from command line, `config.json` is loaded and values are read from it.
- `Ports` and `Targets` are processed from the above file. Targets are processed in a way that a new struct `targetServer` is initialised which implements the `Target` interface.
- The `Target` interface has necessary methods needed for querying remote address information.
- To listen into every port from the config file, a **new goroutine** is created to setup the proxy server.
- `Setup` method basically resolves the local address i.e our port and starts listening to it while checking for any connections.
- To check for new connections it is run in an infinite loop and the proxy server is ran in a different goroutine.
- Loadbalancing is done using a simple **roundrobin** algorithm. It checks for active host and if the host is not present, it moves on to a different host i.e remote address.
- After active remote address is found, a new TCP connection is made via proxy.
- Bidirectional copy of data is done, so that client and server can both see the response sent from the remote address.

## eBPF

- BPF kernel code resides in `bpf/bpf_proxy_dispatch.c`
- There are 2 maps defined `proxy_ports` of type `BPF_MAP_TYPE_HASH` and `proxy_sock` of type `BPF_MAP_TYPE_SOCKMAP`.
- `proxy_ports` map store information about the ports that are inside the `config.json`
- `proxy_sock` store the information about the `fd` of the `proxifie's` listerner socket.
- `SEC` contains that code that will be loaded into the memory.
- `bpf_prog` is the program name that get's attached to the kernel.
- port number is found by `ctx->local_port` and if the port is not inside the `proxy_ports` map the socket connection is passed forward.
- `sk` contains the information about `proxy_sock` map's 0'th element which is our `fd` of `proxifie's` listerner socket.
- `bpf_sk_assign` assgns the `ctx` to our socket `sk`
- Package `tracer` has all the information how to load eBPF program.
- Once the proxify application is run by `sudo ./proxify -b`. Userspace side loads the bpf ELF file `bpf_proxy_dispatch.o`. The function `NewTracer` inside `pkg/tracer/tracer.go` helps in achieving that.
- After that userspace loads both the maps, initialise and populate them with respective values and creates a link between network ns and eBPF program. This is done by `RegisterIface` function.
