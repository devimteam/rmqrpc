package main

import (
	"context"
	"fmt"

	"github.com/devimteam/rmqrpc/json"
	"github.com/streadway/amqp"
)

type RequestTest struct {
	Name string
}

type ResponseTest struct {
	Name string
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

	c := json.NewClient(ch)

	out, err := c.Invoke(ctx, "Test.MakeString", RequestTest{Name: "Pegas2"}, false, ResponseTest{})

	if err != nil {
		panic(err)
	}

	val := <-out

	fmt.Println(val)
}
