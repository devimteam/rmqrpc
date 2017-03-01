package main

import (
	"fmt"

	pbmain "github.com/l-vitaly/rmqrpc/examples/proto_client/pb"
	"github.com/l-vitaly/rmqrpc/metadata"
	"github.com/l-vitaly/rmqrpc/proto"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

func main() {
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

	ctx := context.Background()
	c := proto.NewClient(ch)

	ctx = metadata.NewContext(ctx, metadata.New(map[string]string{"auth": "token"}))

	out, err := c.Invoke(ctx, "test.MakeString", &pbmain.TestRequest{Name: "Pegas2"}, true, pbmain.TestResponse{})

	if err != nil {
		panic(err)
	}

	val := <-out

	fmt.Println(val.(*pbmain.TestResponse).Name)
}
