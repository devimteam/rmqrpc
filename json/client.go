package json

import (
	"encoding/json"
	"reflect"

	"github.com/devimteam/rmqrpc"
	"github.com/streadway/amqp"
)

func encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decode(b []byte, reply interface{}) (interface{}, error) {
	var resp serverResponse
	json.Unmarshal(b, &resp)

	if resp.Error != nil {
		return nil, resp.Error
	}

	newReply := reflect.New(
		reflect.TypeOf(
			reflect.Indirect(reflect.ValueOf(reply)).Interface(),
		),
	)
	err := json.Unmarshal(resp.Result, newReply.Interface())
	if err != nil {
		return nil, err
	}
	return newReply.Interface(), nil
}

// NewClient returns a new RMQ RPC client.
func NewClient(ch *amqp.Channel) rmqrpc.Client {
	return rmqrpc.NewClient(ch, "application/json", encode, decode)
}
