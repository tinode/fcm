# Golang FCM
Basic implementation of FCM (firebase cloud messaging) in Go. Only HTTP requests with JSON payload are supported. The reason for this implementation is to ensure HTTP requests are properly pooled and responses are closed.
This library is used by https://github.com/tinode/chat.

## Documentation

[https://godoc.org/github.com/tinode/fcm](https://godoc.org/github.com/tinode/fcm)

## Usage:

```
  client := fcm.NewClient(your_fcm_api_key)

  message := &fcm.HttpMessage{...initialize your message...}
  response := client.SendHttp(message)
```

The `client` is safe to use from multiple go routines at the same time. The client maintains a pool of HTTP connections. It recycles them as needed. Do not recreate client for every request because it's wasteful.
`SendHttp` is a blocking call. 

Sample code: https://github.com/tinode/chat/blob/master/server/push/fcm/push_fcm.go

## Installation

```
go get github.com/tinode/fcm
```
