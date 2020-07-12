# Log Streamer
this is a simple client/server utility designed for streaming logs over Standard HTTP connections

supports more then one named channel

# Building it
```
go build
```

# Using it 
## start the server
```
logstreamer server :8080
```
## start a writer
{some long running task } |logstreamer write http://localhost:8080/log/channelName
any text written to stdin will be appended to the output received by the reader
### example
```
tail -f logfile.log | logstreamer write http://localhost:8080/log/channelName
```
## start a reader
the reader can be any http client that supports Chunked Transfer for example your web Browser or curl. there is also a built in reader command, if you don't have an existing suitable client.

text written by the writer will be received by the reader in realtime.

the reader connection will be closed when the writer disconnects
### Example
```
curl http://localhost:8080/log/ChannelName
```
or
```
logstreamer read http://localhost:8080/log/ChannelName
```

