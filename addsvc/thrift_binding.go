package main

import (
	"time"

	"golang.org/x/net/context"

	thriftadd "github.com/peterbourgon/gokit/addsvc/thrift/gen-go/add"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/server"
)

// A binding wraps an Endpoint so that it's usable by a transport.
// thriftBinding makes an Endpoint usable over Thrift.
type thriftBinding struct {
	context.Context
	server.Endpoint
}

// Add implements Thrift's AddService interface.
func (tb thriftBinding) Add(a, b int64) (*thriftadd.AddReply, error) {
	r, err := tb.Endpoint(tb.Context, request{a, b})
	if err != nil {
		return nil, err
	}

	resp, ok := r.(*response)
	if !ok {
		return nil, server.ErrBadCast
	}

	return &thriftadd.AddReply{Value: resp.V}, nil
}

func thriftInstrument(requests metrics.Counter, duration metrics.Histogram, next thriftadd.AddService) thriftadd.AddService {
	return thriftInstrumented{requests, duration, next}
}

type thriftInstrumented struct {
	requests metrics.Counter
	duration metrics.Histogram
	next     thriftadd.AddService
}

func (i thriftInstrumented) Add(a, b int64) (*thriftadd.AddReply, error) {
	i.requests.Add(1)
	defer func(begin time.Time) { i.duration.Observe(time.Since(begin).Nanoseconds()) }(time.Now())
	return i.next.Add(a, b)
}