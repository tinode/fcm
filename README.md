# fcm
Basic implementation of FCM (firebase cloud messaging) in Go. Only HTTP requests with JSON payload are supported.

Usage:

  client := fcm.NewClient(your_fcm_api_key)

  message := &fcm.HttpMessage{...initialize your message...}
  response := client.Send(message)

The `client` is safe to use from multiple go routines at the same time. Do not recreate client for every request because it's wasteful.
`Send` is a blocking call. 