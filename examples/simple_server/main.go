package main

import (
	"fmt"

	"github.com/devimteam/rmqrpc"
	"github.com/devimteam/rmqrpc/json"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type RequestTest struct {
	Name string
}

type ResponseTest struct {
	Name string
}

type testService struct {
}

func (*testService) MakeString(ctx context.Context, req *RequestTest) (*ResponseTest, error) {
	return &ResponseTest{req.Name}, nil
}

func main() {
	ctx := context.Background()

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s", "localhost:32769"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	s := rmqrpc.NewServer(ch, ctx, 5) // 5 workers for method
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(testService), "Test")
	s.Listen()
}
