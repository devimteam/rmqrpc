package rmqrpc

// ErrorCode JSON RPC error code type
type ErrorCode int64

const (
    // ErrParse Invalid JSON was received by the server.
    ErrParse ErrorCode = -32700
    // ErrInvalidRequest The JSON sent is not a valid Request object.
    ErrInvalidRequest ErrorCode = -32600
    // ErrMethodNotFound The method does not exist / is not available.
    ErrMethodNotFound ErrorCode = -32601
    // ErrBadParams Invalid method parameter(s).
    ErrBadParams ErrorCode = -32602
    // ErrInternal Internal JSON-RPC error.
    ErrInternal ErrorCode = -32603
    // ErrServer Reserved for implementation-defined server-errors.
    ErrServer ErrorCode = -32000
)
