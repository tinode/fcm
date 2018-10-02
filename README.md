# Golang FCM

Basic implementation of FCM (firebase cloud messaging) in Go. Only HTTP requests with JSON payload are supported.
This package uses legacy HTTP API (pre-v1 API). 

## Documentation

[https://godoc.org/github.com/tinode/fcm](https://godoc.org/github.com/tinode/fcm)

## Usage:

```
  client := fcm.NewClient(your_fcm_api_key)

  message := &fcm.HttpMessage{...initialize your message...}
  response := client.SendHttp(message)
```

The `client` is safe to use from multiple go routines at the same time. Do not recreate client for every request because it's wasteful.
`SendHttp` is a blocking call. 

## Installation

```
go get github.com/tinode/fcm
```