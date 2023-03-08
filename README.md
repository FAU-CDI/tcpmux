# tcpmux

TCPMux is a tool that can multiplex http, https and ssh connections on a single port.
It makes uses of [cmux](https://github.com/soheilhy/cmux). 

It is automatically deployed as a GitHub Package on every commit.

To use it, run:

`docker run --read-only -p 8000:8000 ghcr.io/fau-cdi/tcpmux:latest [...args]`