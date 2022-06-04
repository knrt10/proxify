package proxy

import (
	"io"
	"log"
	"net"
)

type proxy struct {
	laddr, raddr           *net.TCPAddr
	lconn, rconn           io.ReadWriteCloser
	sentData, receivedData string
}

// Setup checks localaddress i.e port for a valid TCP conncetion and
// then starts a new proxy instance
func Setup(localAddr string, lb *LoadBalancer) *proxy {
	laddr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		log.Fatalln("Failed to resolve local address:", err)
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatalln("Failed to open local port to listen:", err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalln("Failed to accept connection:", err)
			continue
		}

		// Get remote address of an alive host
		remoteAddr := lb.GetNextAvailableTarget().Address()
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		if err != nil {
			log.Fatalln("Failed to resolve remote address:", err)
		}

		var p *proxy
		p = p.new(conn, laddr, raddr)
		go p.start()
	}
}

// new create a new Proxy instance.
func (p *proxy) new(lconn *net.TCPConn, laddr, raddr *net.TCPAddr) *proxy {
	return &proxy{
		lconn: lconn,
		laddr: laddr,
		raddr: raddr,
	}
}

// starts opens connection to remote and start proxying data.
func (p *proxy) start() {
	defer p.lconn.Close()

	var err error
	p.rconn, err = net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		log.Println("Remote connection failed:", err)
		return
	}
	defer p.rconn.Close()

	//bidirectional copy of data
	closer := make(chan struct{}, 2)
	go p.copy(closer, p.lconn, p.rconn)
	go p.copy(closer, p.rconn, p.lconn)
	<-closer
}

func (p *proxy) copy(closer chan struct{}, src, dst io.ReadWriter) {
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}

		b := buff[:n]
		_, err = dst.Write(b)
		if err != nil {
			return
		}

		if src == p.lconn {
			p.sentData = string(b)
			log.Println("Sent data from client:", p.sentData)
		} else {
			p.receivedData = string(b)
			log.Println("Received data from server:", p.receivedData)
			closer <- struct{}{}
		}
	}
}
