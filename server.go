package rmqrpc

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/l-vitaly/rmqrpc/metadata"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

var (
	ErrCodecRequired = errors.New("codec is required")
)

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(Request) CodecRequest
}

// CodecRequest decodes a request and encodes a response using a specific
// serialization scheme.
type CodecRequest interface {
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(ResponseWriter, interface{})
	// Writes an error produced by the server.
	WriteError(ResponseWriter, error)
}

// Server serves registered RMQ RPC services using registered codecs.
type Server struct {
	ctx        context.Context
	ch         *amqp.Channel
	codecs     map[string]Codec
	serviceMap *serviceMap
	workers    int
}

// NewServer returns a new RMQ RPC server.
func NewServer(ch *amqp.Channel, ctx context.Context, workers int) *Server {
	return &Server{
		ch:         ch,
		ctx:        ctx,
		codecs:     make(map[string]Codec),
		serviceMap: new(serviceMap),
		workers:    workers,
	}
}

func (s *Server) RegisterCodec(codec Codec, contentType string) {
	s.codecs[strings.ToLower(contentType)] = codec
}

func (s *Server) RegisterService(receiver interface{}, name string) error {
	return s.serviceMap.register(receiver, name)
}

func (s *Server) serveRMQ(rw ResponseWriter, r Request, md metadata.MD, serviceSpec *service, methodSpec *serviceMethod) {
	contentType := r.ContentType()

	ctx := metadata.NewContext(s.ctx, md)

	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}

	var codec Codec

	if contentType == "" && len(s.codecs) == 1 {
		// If Content-Type is not set and only one codec has been registered,
		// then default to that codec.
		for _, c := range s.codecs {
			codec = c
		}
	} else if codec = s.codecs[strings.ToLower(contentType)]; codec == nil {
		log.Println(errors.New("rpc: unrecognized Content-Type: " + contentType))
		rw.Commit()
		return
	}

	// Create a new codec request.
	codecReq := codec.NewRequest(r)

	argReq := reflect.New(methodSpec.argReqType)
	if errRead := codecReq.ReadRequest(argReq.Interface()); errRead != nil {
		codecReq.WriteError(rw, errRead)
		rw.Commit()
		return
	}

	retValues := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
		argReq,
	})

	// Cast the result to error if needed.
	var errResult error
	errInter := retValues[1].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	if errResult == nil {
		rw.Commit()
	}

	// Encode the response.
	if errResult == nil {
		valRet := retValues[0].Interface()
		codecReq.WriteResponse(rw, valRet)
	} else {
		codecReq.WriteError(rw, errResult)
	}
}

func (s *Server) serviceWorker(
	deliveryCh <-chan amqp.Delivery, serviceSpec *service, methodSpec *serviceMethod,
) {
	go func() {
		for d := range deliveryCh {
			rw := &responseWriter{ch: s.ch, d: d}
			r := &request{d: d}

			md := metadata.MD{}
			for k, v := range d.Headers {
				md[k] = v.(string)
			}
			s.serveRMQ(rw, r, md, serviceSpec, methodSpec)
		}
	}()
}

func (s *Server) Listen() error {
	if len(s.codecs) == 0 {
		return ErrCodecRequired
	}

	waiting := make(chan bool)

	for srvName, serviceSpec := range s.serviceMap.services {
		for srvMethodName, methodSpec := range serviceSpec.methods {
			q, err := s.ch.QueueDeclare(
				srvName+"."+srvMethodName,
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				return err
			}
			err = s.ch.Qos(
				1,
				0,
				false,
			)
			if err != nil {
				return err
			}
			deliveryCh, err := s.ch.Consume(
				q.Name,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			for i := 0; i < s.workers; i++ {
				s.serviceWorker(deliveryCh, serviceSpec, methodSpec)
			}
		}
	}

	<-waiting

	return nil
}
