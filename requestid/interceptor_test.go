package requestid

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/transport"
)

type testRequest struct{}

type testResponse struct{}

func DummyContextWithServerTransportStream() context.Context {
	expectedStream := &transport.Stream{}
	return grpc.NewContextWithServerTransportStream(context.Background(), expectedStream)
}

func TestUnaryServerInterceptorWithoutRequestId(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqID, exists := FromContext(ctx)
		if exists && reqID == "" {
			t.Errorf("requestId must be generated by interceptor")
		}
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	_, err := UnaryServerInterceptor()(ctx, testRequest{}, nil, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnaryServerInterceptorWithDummyRequestId(t *testing.T) {
	dummyRequestID := newRequestID()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqID, exists := FromContext(ctx)
		if !exists || reqID != dummyRequestID {
			t.Errorf("expected requestID: %q, returned requestId: %q", dummyRequestID, reqID)
		}
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	md := metadata.Pairs(DefaultRequestIDKey, dummyRequestID)
	newCtx := metadata.NewIncomingContext(ctx, md)
	_, err := UnaryServerInterceptor()(newCtx, testRequest{}, nil, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnaryServerInterceptorPanic(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("TestUnaryServerInterceptorPanic raised panic")
	}
	ctx := DummyContextWithServerTransportStream()
	_, err := UnaryServerInterceptor()(ctx, testRequest{}, nil, handler)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
