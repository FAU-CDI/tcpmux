# tcpmux

TCPMux is a tool that can multiplex http, https and ssh connections on a single port.
It makes uses of [cmux](https://github.com/soheilhy/cmux).

It also optionally supports the [Proxy Protocol](https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt) using [go-proxyproto](https://github.com/pires/go-proxyproto?tab=readme-ov-file)

It is automatically deployed as a GitHub Package on every commit.

To use it, run:

`docker run --read-only -p 8000:8000 ghcr.io/fau-cdi/tcpmux:latest [...args]`