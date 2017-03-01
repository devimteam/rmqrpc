package json

import (
	"encoding/json"
	"github.com/l-vitaly/rmqrpc"
)

type serverResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
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
	if err := json.Unmarshal(c.request.Body(), args); err != nil {
		c.err = &Error{
			Code:    rmqrpc.ErrInvalidRequest,
			Message: err.Error(),
		}
	}
	return c.err
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
func (c *CodecRequest) WriteResponse(w rmqrpc.ResponseWriter, reply interface{}) {
	data, _ := json.Marshal(reply)

	res := &serverResponse{
		Result: data,
	}
	c.writeServerResponse(w, res)
}

// WriteError send error response.
func (c *CodecRequest) WriteError(w rmqrpc.ResponseWriter, err error) {
	jsonErr, ok := err.(*Error)

	if !ok {
		jsonErr = &Error{
			Code:    rmqrpc.ErrInvalidRequest,
			Message: err.Error(),
		}
	}

	res := &serverResponse{
		Error: jsonErr,
	}

	c.writeServerResponse(w, res)
}

func (c *CodecRequest) writeServerResponse(w rmqrpc.ResponseWriter, res *serverResponse) {
	resp, _ := json.Marshal(res)
	w.SetContentType("application/json")
	w.Write(resp)
}
