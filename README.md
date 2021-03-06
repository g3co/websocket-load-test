Testing websocket servers.

Install the command line tool first.

```go get github.com/g3co/websocket-load-test```

Then you will have access to the application. Exec command below 
for open simultaneous connections to the WS server. The application 
just keeping connections, reading messages from WS, and counting 
the total amount of messages.

```websocket-load-test [url] [connection quantity]```