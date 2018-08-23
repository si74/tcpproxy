# tcpproxy - layer 4

## Background

- IPVS - TCP loadbalancer built into Linux kernel
- goal is to build something similar to IPVS [https://github.com/torvalds/linux/tree/master/net/netfilter/ipvs](https://github.com/torvalds/linux/tree/master/net/netfilter/ipvs)

## Considerations

- Focus on supporting IPv4
- 1 front and many backends for now
- every client connection coming to proxy requires opening a backend connection
- This is a TCP proxy - uses TCP handshake to establish a connection (SYN, SYN-ACK, ACK).
- Load balancing strategies - round robin, random, weighted least connections, same client to same backend, etc.
  - Source: [http://www.linuxvirtualserver.org/docs/scheduling.html](http://www.linuxvirtualserver.org/docs/scheduling.html)
- How to close connections in TCP - orderly way (FIN, and ACK. (Tidbit: TCP connections can be half-open.)
or abruptly with RST flag ("connection reset by peer.")
- Max number of connections or go-routine limitation (if 1 per connection).
- May have keep-alive - send ACK flag to keep connections open. But every incoming connection (even if from same client), don't necessarily want to reuse back-end connection. Otherwise could mix data on backend connection. 

## Pseudocode:

1. Configuration - either via flags or file.
2. Tell operating system that service is listening on a local IP address and port. (transport layer - 4)
3. Blocks until we accept a connection to the listener. Output is conn (i.e. socket). Runs in loop.
4. In a new go-routine, do the following:
   - Open a connection to a backend using some strategy.
   - If there in an error (after retrying), close incoming client connection.
   - If successful, open 2 new go-routines - 1 for shuttling packets from client to backend and 1 for
   backend to client.
     -  Each go-routine does this until there is an error (implying the connection is dead), do following in loop:
        - Reads bytes from connection A and writes bytes to connection B.
   - When error happens in either of these go-routines, shut down main go-routine.
5. Just wait for the next incoming connection to listener.

## How to test:
1. Leveraging example.com and telnet //TODO(sneha): add more details on how to do this with the host header

```
$ telnet localhost 9090
Trying ::1...
telnet: connect to address ::1: Connection refused
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
GET / HTTP/1.1
Host: example.com
```

2. Leveraging netcat and telnet

```
$ nc -l 12345
hi
```

```
telnet localhost 9090
Trying ::1...
telnet: connect to address ::1: Connection refused
Trying 127.0.0.1...
```
