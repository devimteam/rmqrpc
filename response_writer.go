package rmqrpc

import "github.com/streadway/amqp"

type ResponseWriter interface {
	Write([]byte)
	Commit()
	SetContentType(string)
}

type responseWriter struct {
	d           amqp.Delivery
	ch          *amqp.Channel
	contentType string
}

func (rw *responseWriter) Write(content []byte) {
	msg := amqp.Publishing{
		ContentType:   rw.contentType,
		CorrelationId: rw.d.CorrelationId,
		Body:          content,
	}
	rw.ch.Publish("", rw.d.ReplyTo, false, false, msg)
}

func (rw *responseWriter) SetContentType(contentType string) {
	rw.contentType = contentType
}

func (rw *responseWriter) Commit() {
	rw.d.Ack(false)
}
