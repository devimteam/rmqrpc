package proto

import (
	"github.com/gogo/protobuf/proto"
	"github.com/l-vitaly/rmqrpc"
)

func (e *Error) Error() string {
	return e.Message
}

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) NewRequest(r rmqrpc.Request) rmqrpc.CodecRequest {
	return &CodecRequest{request: r, err: nil}
}

type CodecRequest struct {
	request rmqrpc.Request
	err     error
}

func (c *CodecRequest) ReadRequest(args interface{}) error {
    if err := proto.Unmarshal(c.request.Body(), args.(proto.Message)); err != nil {
		c.err = &Error{
			Code:    int64(rmqrpc.ErrInvalidRequest),
			Message: err.Error(),
		}
	}
	return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
func (c *CodecRequest) WriteResponse(w rmqrpc.ResponseWriter, reply interface{}) {
	data, _ := proto.Marshal(reply.(proto.Message))
	res := &ServerResponse{
		Result: data,
	}
	c.writeServerResponse(w, res)
}

// WriteError send error response.
func (c *CodecRequest) WriteError(w rmqrpc.ResponseWriter, err error) {
	rpcErr, ok := err.(*Error)

	var protoErr *Error

	if !ok {
		protoErr = &Error{
			Code:    int64(rmqrpc.ErrInvalidRequest),
			Message: err.Error(),
		}
	} else {
		protoErr = &Error{
			Code:    int64(rpcErr.Code),
			Message: rpcErr.Message,
		}
	}

	res := &ServerResponse{
		Error: protoErr,
	}

	c.writeServerResponse(w, res)
}

func (c *CodecRequest) writeServerResponse(w rmqrpc.ResponseWriter, res *ServerResponse) {
	resp, _ := proto.Marshal(res)
	w.SetContentType("application/protobuf")
	w.Write(resp)
}
