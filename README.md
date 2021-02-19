# mergectx

This pkg implements a context which can be used globally in a server. Incoming request contexts can then be merged with the global context.
On cancellation of the global or a request context the merged context gets also canceled.
This is especially useful for canceling log running request before starting a graceful server shutdown.

A more concrete example here is when using grpc streams then such a context can be useful on the grpc server.
