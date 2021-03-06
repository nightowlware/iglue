# iglue
Pronounced "igloo".
Stands for "ipc glue" or "inter-process glue". 

It's a very lightweight ipc library, written in Go. All inter-process messages are fixed size strings, 
exposed to the process via a Go channel.

For an example on how to use this library, see the example folder.

## Design
iglue uses named pipes/fifos as the underlying ipc mechanism. 

"iglue IDs" map directly to the names of the pipes/fifos on the system.

For sending, a fixed-size string message is written directly to the pipe matching the destination iglue ID.  
For receiving, a dedicated goroutine constantly waits for messages on the named pipe; when a new message arrives,
it is written to the client's receiving channel.
