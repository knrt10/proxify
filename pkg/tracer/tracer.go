package tracer

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/knrt10/proxify/pkg/assets"
	"golang.org/x/sys/unix"
)

import "C"

type iface struct {
	ifindex    int
	coll       *ebpf.Collection
	proSocMap  *ebpf.Map
	proPortMap *ebpf.Map
	sockFd     int
	prog       *ebpf.Program
}

type Tracer struct {
	spec  *ebpf.CollectionSpec
	ports []int
	iface *iface
}

// NewTracer starts a new eBPF tracer, which loads ELF file and reads it
func NewTracer(ports []int) (*Tracer, error) {
	// This is done to increase value of MEMLOCK
	if err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	}); err != nil {
		return nil, fmt.Errorf("cannot set rlimit: %w", err)
	}

	t := &Tracer{
		ports: ports,
	}

	// Load the eBPF collection from the ELF file
	asset, err := assets.Asset("bpf_proxy_dispatch.o")
	if err != nil {
		return nil, fmt.Errorf("cannot open asset: %w", err)
	}

	t.spec, err = ebpf.LoadCollectionSpecFromReader(bytes.NewReader(asset[:]))
	if err != nil {
		return nil, fmt.Errorf("cannot load asset: %w", err)
	}

	return t, nil
}

// initPortsMap is used to populate port values in proxy_ports map
func (t *Tracer) initPortsMap(m *ebpf.Map) error {
	for _, port := range t.ports {
		key := uint32(port)
		value := 0
		if err := m.Put(unsafe.Pointer(&key), unsafe.Pointer(&value)); err != nil {
			return err
		}
	}
	return nil
}

// initSocMap is used to put socket value of port 8080 in proxy_sock map
func (t *Tracer) initSocMap(m *ebpf.Map, sockFd int) error {
	key := 0
	if err := m.Put(unsafe.Pointer(&key), unsafe.Pointer(&sockFd)); err != nil {
		return err
	}

	return nil
}

/** RegisterIface registers a new interface on every run of the program
	- it creates new eBPF collection
	- initialize and populate proxy_ports and proxy_sock map
	- get's the eBPF program
	- creates a link between network ns and eBPF program
**/
func (t *Tracer) RegisterIface(ifindex int, sockFd int) (err error) {
	i := &iface{
		ifindex: ifindex,
		sockFd:  sockFd,
	}

	defer func() {
		if err != nil {
			closeIface(i)
		}
	}()

	i.coll, err = ebpf.NewCollection(t.spec)
	if err != nil {
		return fmt.Errorf("cannot create new ebpf collection: %s", err)
	}

	var ok bool
	i.proPortMap, ok = i.coll.Maps["proxy_ports"]
	if !ok {
		return fmt.Errorf("no map named proxy_ports found")
	}

	if err = t.initPortsMap(i.proPortMap); err != nil {
		return fmt.Errorf("cannot initialize proxy_ports map: %s:", err)
	}

	i.proSocMap, ok = i.coll.Maps["proxy_sock"]
	if !ok {
		return fmt.Errorf("no map named proxy_sock found")
	}

	if err = t.initSocMap(i.proSocMap, i.sockFd); err != nil {
		return fmt.Errorf("cannot initialize proxy_sock map: %s:", err)
	}

	i.prog, ok = i.coll.Programs["bpf_prog"]
	if !ok {
		return fmt.Errorf("bpf program not found")
	}

	// Get an FD for this process network namespace (netns)
	file, err := os.Open("/proc/self/ns/net")
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}

	fd := file.Fd()

	if _, err = link.AttachRawLink(link.RawLinkOptions{
		Target:  int(fd),
		Program: i.prog,
		Attach:  ebpf.AttachSkLookup,
		Flags:   0,
	}); err != nil {
		return fmt.Errorf("unable to attach program to network namespace: %s", err)
	}

	t.iface = i

	return nil
}

func closeIface(i *iface) {
	if i.coll != nil {
		i.coll.Close()
	}
	if i.sockFd != -1 {
		syscall.Close(i.sockFd)
	}
}
