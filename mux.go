package tcpmux

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"github.com/soheilhy/cmux"
)

//go:generate gogenlicense -m

// Mux represents a multiplexer that can forward
type Mux struct {
	Logger *log.Logger
}

type Target struct {
	HTTP string
	TLS  string
	Rest string
}

// Serve starts serving the provided listener until the context is closed.
func (mux *Mux) Serve(ctx context.Context, l net.Listener, target Target) {
	m := cmux.New(l)

	var wg sync.WaitGroup
	var listeners []net.Listener

	if target.HTTP != "" {
		l := m.Match(cmux.HTTP1(), cmux.HTTP2())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, l, target.HTTP)
	}

	if target.TLS != "" {
		l := m.Match(cmux.TLS())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, l, target.TLS)
	}

	if target.Rest != "" {
		l := m.Match(cmux.Any())
		listeners = append(listeners, l)

		mux.forwardTask(&wg, l, target.Rest)
	}

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

func (mux *Mux) forwardTask(wg *sync.WaitGroup, l net.Listener, remote string) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		mux.Logger.Printf("forwarding to %s\n", remote)

		for {
			conn, err := l.Accept()
			if err != nil {
				mux.Logger.Printf("forward.Accept(%s): returned err=%s\n", remote, err)
				return
			}
			wg.Add(1)
			go mux.forward(wg, conn, remote)
		}
	}()
}

func (mux *Mux) forward(wg *sync.WaitGroup, src net.Conn, remote string) error {
	defer wg.Done()

	dst, err := net.Dial("tcp", remote)
	if err != nil {
		mux.Logger.Printf("forward.Dial(%s): returned err=%s\n", remote, err)
		return err
	}
	mux.Logger.Printf("forward.Accept(%s): ok\n", remote)

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
