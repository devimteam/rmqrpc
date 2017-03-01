package main

import (
	"fmt"

	"github.com/l-vitaly/rmqrpc"
	pbmain "github.com/l-vitaly/rmqrpc/examples/proto_server/pb"
	"github.com/l-vitaly/rmqrpc/metadata"
	"github.com/l-vitaly/rmqrpc/proto"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type testService struct {
}

func (*testService) MakeString(ctx context.Context, req *pbmain.TestRequest) (*pbmain.TestResponse, error) {
	md, _ := metadata.FromContext(ctx)

	fmt.Println(md)

	return &pbmain.TestResponse{Name: req.Name}, nil
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

	s.RegisterCodec(proto.NewCodec(), "application/protobuf")
	s.RegisterService(new(testService), "test")

	s.Listen()
}
