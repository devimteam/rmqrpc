package json

import "github.com/devimteam/rmqrpc"

type Error struct {
	// A Number that indicates the error type that occurred.
	Code rmqrpc.ErrorCode `json:"code"`
	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"`
	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
