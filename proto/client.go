package proto

import (
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/devimteam/rmqrpc"
	"github.com/streadway/amqp"
)

func encode(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func decode(b []byte, reply interface{}) (interface{}, error) {
	resp := &ServerResponse{}
	err := proto.Unmarshal(b, resp)
    if err != nil {
        return nil, err
    }

	if resp.Error != nil {
		return nil, resp.Error
	}

	newReply := reflect.New(
		reflect.TypeOf(
			reflect.Indirect(reflect.ValueOf(reply)).Interface(),
		),
	)

	err = proto.Unmarshal(resp.Result, newReply.Interface().(proto.Message))
	if err != nil {
		return nil, err
	}
	return newReply.Interface(), nil
}

// NewClient returns a new RMQ RPC client.
func NewClient(ch *amqp.Channel) rmqrpc.Client {
	return rmqrpc.NewClient(ch, "application/protobuf", encode, decode)
}
