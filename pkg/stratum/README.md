# Stratum

A portable library to manage a stratum client or server. This was created to isolate
the more challenging (asynchronous) logic of TCP clients and servers, so that
they can be tested separately from heavily asynchronous client and server implementations.
Though the client is more "useful" since it has more functionality, the server is still
extremely valuable to have a plug and play TCP listener with connection management. For
both the client and the server, TLS support will eventually be added.

The client is fairly generalized, only with some bias on the requirement of 
the handshake request (usually `mining.subscribe`, or `eth_submitLogin`). It automatically 
manages reconnects and JSON RPC request ID management, so that it can properly associate 
requests and their responses asynchronously. The client is completely thread-safe and 
generally used in that way. It does not, however, manage multi-host rotation and
prioritization, since that is the purpose of the `TCPPool` in `pkg/hostpool`. The
only thing missing is probably batch request and multiple (or sequential) handshake 
request support (like `mining.subscribe` and `mining.authorize`).

The server is a little bit more generalized since the server `Conn` has some specific
needs with data store in the client. Currently this is managed through getters and setters,
though it *could* become an interface that the user implements if this becomes too
cumbersome. For all current needs, it works well. The server implements no routing and leaves
that up to the user, it is merely meant to be a convenient wrapper for `net.Listener` and the 
corresponding list of `net.Conn`s. One thing that could be implemented is better backoff
for disconnecting clients when the server is shut down instead of disconnecting them
all at once (which can have adverse affects for the clients in production).
