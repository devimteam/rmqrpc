package rmqrpc

import (
	"context"
	"math/rand"

	"github.com/devimteam/rmqrpc/metadata"
	"github.com/streadway/amqp"
)

type EncodeFunc func(interface{}) ([]byte, error)
type DecodeFunc func([]byte, interface{}) (interface{}, error)

type Client interface {
	Invoke(ctx context.Context, method string, req interface{}, async bool, reply interface{}) (chan interface{}, error)
}

// Client client RMQ RPC services.
type client struct {
	ch          *amqp.Channel
	contentType string
	enc         EncodeFunc
	dec         DecodeFunc
}

// NewClient returns a new RMQ RPC Client.
func NewClient(ch *amqp.Channel, contentType string, enc EncodeFunc, dec DecodeFunc) Client {
	return &client{
		ch:          ch,
		contentType: contentType,
		enc:         enc,
		dec:         dec,
	}
}

func (c *client) Invoke(ctx context.Context, method string, req interface{}, async bool, reply interface{}) (chan interface{}, error) {
	q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	msgs, err := c.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	corrId := c.makeCorrelationId(32)

	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	headers := make(map[string]interface{})
	for k, v := range md {
		headers[k] = v
	}

	reqRaw, err := c.enc(req)
	if err != nil {
		return nil, err
	}

	err = c.ch.Publish(
		"",
		method,
		false,
		false,
		amqp.Publishing{
			Headers:       headers,
			ContentType:   c.contentType,
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          reqRaw,
		})
	if err != nil {
		return nil, err
	}

	out := make(chan interface{})
	if async {
		go c.replyWait(corrId, msgs, out, reply)
	} else {
		close(out)
	}
	return out, nil
}

func (c *client) makeCorrelationId(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(c.randInt(65, 90))
	}
	return string(bytes)
}

func (c *client) randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func (c *client) replyWait(corrId string, msgs <-chan amqp.Delivery, out chan interface{}, reply interface{}) {
	for d := range msgs {
		if corrId == d.CorrelationId {
			res, err := c.dec(d.Body, reply)
			if err != nil {
				out <- err
			} else {
				out <- res
			}
			break
		}
	}
	defer close(out)
}
