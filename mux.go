package tcpmux

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"github.com/pires/go-proxyproto"
	"github.com/soheilhy/cmux"
)

//go:generate go tool gogenlicense -m

// Mux represents a multiplexer that can forward
type Mux struct {
	Logger *log.Logger
}

type Target struct {
	HTTP              string
	HTTPProxyProtocol bool

	TLS              string
	TLSProxyProtocol bool

	Rest              string
	RestProxyProtocol bool
}

// Serve starts serving the provided listener until the context is closed.
func (mux *Mux) Serve(ctx context.Context, l net.Listener, target Target) {
	m := cmux.New(l)

	var wg sync.WaitGroup
	var listeners []net.Listener

	if target.HTTP != "" {
		l := m.Match(cmux.HTTP1(), cmux.HTTP2())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, "HTTP", l, target.HTTP, target.HTTPProxyProtocol)
	}

	if target.TLS != "" {
		l := m.Match(cmux.TLS())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, "TLS", l, target.TLS, target.TLSProxyProtocol)
	}

	if target.Rest != "" {
		l := m.Match(cmux.Any())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, "Rest", l, target.Rest, target.RestProxyProtocol)
	}

	go m.Serve()

	<-ctx.Done()
	log.Println("Stopping")
	for _, l := range listeners {
		l.Close()
	}

	// wait for graceful exit
	wg.Wait()
}

func New(logger *log.Logger) *Mux {
	if logger == nil {
		logger = log.Default()
	}
	return &Mux{Logger: logger}
}

func (mux *Mux) forwardTask(wg *sync.WaitGroup, name string, l net.Listener, remote string, proxyProtocol bool) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		mux.Logger.Printf("forwarding %s to %s (proxy protocol: %v)\n", name, remote, proxyProtocol)

		for {
			conn, err := l.Accept()
			if err != nil {
				mux.Logger.Printf("forward.Accept(%s): returned err=%s\n", remote, err)
				return
			}
			wg.Add(1)
			go mux.forward(wg, conn, remote, proxyProtocol)
		}
	}()
}

func (mux *Mux) forward(wg *sync.WaitGroup, src net.Conn, remote string, proxyProtocol bool) error {
	defer wg.Done()

	dst, err := net.Dial("tcp", remote)
	if err != nil {
		mux.Logger.Printf("forward.Dial(%s): returned err=%s\n", remote, err)
		return err
	}

	if proxyProtocol {
		header := proxyproto.HeaderProxyFromAddrs(2, src.RemoteAddr(), src.LocalAddr())
		if _, err := header.WriteTo(dst); err != nil {
			mux.Logger.Printf("forward.ProxyHeader(%s): returned err=%s\n", remote, err)
			src.Close()
			dst.Close()
			return err
		}
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer src.Close()
		defer dst.Close()

		io.Copy(dst, src)
	}()

	go func() {
		defer wg.Done()
		defer src.Close()
		defer dst.Close()

		io.Copy(src, dst)
	}()

	return nil
}
