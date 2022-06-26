# go-rudp

Reliable UDP implementation written purely in Go, powered by [Pion's SCTP](https://github.com/pion/sctp) and [DCEP](https://github.com/pion/datachannel).

## Package name: `rudp`

## Overview
### Features
* Implements user-space SCTP over UDP
* TCP like client/server architecture
* DCEP support
  - Ordered / unordered
  - Partial reliability

### Goals
* Initial motivation was to test pion/sctp
* Makes it easy to create UDP based applications

### Difference from WebRTC
* Client/Server (not peer-to-peer)
* Simple (Data Channel only)
  - No Signaling
  - No ICE
  - No DTLS
  - No SRTP

### Note
* Best for creating tools
* Production use is NOT recommended

## Examples
(TODO)
> See [sctptest](https://github.com/enobufs/sctptest) for now.

## TODO
* Allocate server resource on CookieEcho (against DDoS)
