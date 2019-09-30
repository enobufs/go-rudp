# go-rudp

Reliable UDP implementation written purely in Go, powered by [Pion's SCTP](https://github.com/pion/sctp).

## Package name: `rudp`

## Overview
### Features
* Implements user-space SCTP over UDP
* TCP like client/server architecture
* DCEP support (TODO)
  - Ordered / unordered
  - Partial reliability

### Goals
* Initial motivation was to test pion/sctp
* Makes it easy to create UDP based applications

### Difference from WebRTC
* Data Channel only
* Simple
  - No Signaling
  - No ICE
  - No DTLS
  - No SRTP

### Note
* Best for creating tools
* Production use is NOT recommended


## Examples
(TODO)

## TODO
* Implement DCEP
* Bug: Allocate server resource on CookieEcho (against DDoS)
