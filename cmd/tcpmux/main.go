package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/fau-cdi/tcpmux"
)

func main() {
	// open a listener
	l, err := net.Listen("tcp", bindAddress)
	log.Printf("Listening on %s", bindAddress)
	if err != nil {
		log.Fatal(err)
	}

	tcpmux.New(log.Default()).Serve(globalContext, l, tcpmux.Target{
		HTTP:              forwardHTTP,
		HTTPProxyProtocol: forwardHTTPProxy,

		TLS:              forwardTLS,
		TLSProxyProtocol: forwardTLSProxy,

		Rest:              forwardRest,
		RestProxyProtocol: forwardRestProxy,
	})
}

var globalContext context.Context

func init() {
	globalContext, _ = signal.NotifyContext(context.Background(), os.Interrupt)
}

var (
	bindAddress string = "0.0.0.0:8000"

	forwardHTTP      string
	forwardHTTPProxy bool

	forwardTLS      string
	forwardTLSProxy bool

	forwardRest      string
	forwardRestProxy bool
)

func init() {
	var legalFlag bool
	flag.BoolVar(&legalFlag, "legal", legalFlag, "print legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Println("This executable contains (unmodified) code from several different go packages. ")
			fmt.Println("Some of these packages require licensing information to be made available to the end user. ")
			fmt.Println(tcpmux.Notices)
			os.Exit(0)
		}
	}()

	defer flag.Parse()

	flag.StringVar(&bindAddress, "bind", bindAddress, "bind to specific address")

	flag.StringVar(&forwardHTTP, "http", forwardHTTP, "forward http1/2 connections to specific address")
	flag.BoolVar(&forwardHTTPProxy, "http-proxy", forwardHTTPProxy, "send proxy protocol header for http connections")

	flag.StringVar(&forwardTLS, "tls", forwardTLS, "forward tls (https) connections to specific address")
	flag.BoolVar(&forwardTLSProxy, "tls-proxy", forwardTLSProxy, "send proxy protocol header for tls connections")

	flag.StringVar(&forwardRest, "rest", forwardRest, "forward remaining connections to specific address")
	flag.BoolVar(&forwardRestProxy, "rest-proxy", forwardRestProxy, "send proxy protocol header for rest connections")
}
