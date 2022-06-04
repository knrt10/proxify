#include <linux/bpf.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#ifndef printt
#define printt(fmt, ...)                                       \
  ({                                                           \
    char ____fmt[] = fmt;                                      \
    bpf_trace_printk(____fmt, sizeof(____fmt), ##__VA_ARGS__); \
  })
#endif

/* Declare BPF maps */

/* List of open proxy service ports. Key is the port number. */
struct bpf_map_def SEC("maps") proxy_ports = {
    .type = BPF_MAP_TYPE_HASH,
    .max_entries = 1024,
    .key_size = sizeof(__u16),
    .value_size = sizeof(__u8),
};

/* Proxy server socket */
struct bpf_map_def SEC("maps") proxy_sock = {
    .type = BPF_MAP_TYPE_SOCKMAP,
    .max_entries = 1,
    .key_size = sizeof(__u32),
    .value_size = sizeof(__u64),
};

/* Dispatcher program for the proxy service */
SEC("sk_lookup/proxy_dispatch")
int bpf_prog(struct bpf_sk_lookup *ctx)
{
  const __u32 zero = 0;
  struct bpf_sock *sk;
  __u16 port;
  __u8 *open;
  long err;

  port = ctx->local_port;
  open = bpf_map_lookup_elem(&proxy_ports, &port);
  if (!open)
    return SK_PASS;

  /* Get proxy server socket */
  sk = bpf_map_lookup_elem(&proxy_sock, &zero);
  if (!sk)
    return SK_DROP;

  err = bpf_sk_assign(ctx, sk, 0);
  bpf_sk_release(sk);
  return err ? SK_DROP : SK_PASS;
}

char _license[] SEC("license") = "GPL";
