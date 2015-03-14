package main

import (
	"time"

	"github.com/peterbourgon/gokit/addsvc/pb"
	"github.com/peterbourgon/gokit/metrics"
	"github.com/peterbourgon/gokit/server"
	"golang.org/x/net/context"
)

// A binding wraps an Endpoint so that it's usable by a transport. grpcBinding
// makes an Endpoint usable over gRPC.
type grpcBinding struct{ server.Endpoint }

// Add implements the proto3 AddServer by forwarding to the wrapped Endpoint.
func (b grpcBinding) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddReply, error) {
	addReq := request{req.A, req.B}
	r, err := b.Endpoint(ctx, addReq)
	if err != nil {
		return nil, err
	}

	resp, ok := r.(*response)
	if !ok {
		return nil, server.ErrBadCast
	}

	return &pb.AddReply{
		V: resp.V,
	}, nil
}

func grpcInstrument(requests metrics.Counter, duration metrics.Histogram, next pb.AddServer) pb.AddServer {
	return grpcInstrumented{requests, duration, next}
}

type grpcInstrumented struct {
	requests metrics.Counter
	duration metrics.Histogram
	next     pb.AddServer
}

func (i grpcInstrumented) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddReply, error) {
	i.requests.Add(1)
	defer func(begin time.Time) { i.duration.Observe(time.Since(begin).Nanoseconds()) }(time.Now())
	return i.next.Add(ctx, req)
}