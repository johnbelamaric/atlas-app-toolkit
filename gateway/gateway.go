package gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	// DefaultServerAddress is the standard gRPC server address that a REST
	// gateway will connect to.
	DefaultServerAddress = ":9090"
)

// Option is a functional option that modifies the REST gateway on
// initialization
type Option func(*gateway)

type registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) (err error)

type gateway struct {
	serverAddress     string
	serverDialOptions []grpc.DialOption
	endpoints         map[string][]registerFunc
	mux               *http.ServeMux
	gatewayMuxOptions []runtime.ServeMuxOption
}

// NewGateway creates a gRPC REST gateway with HTTP handlers that have been
// generated by the gRPC gateway protoc plugin
func NewGateway(options ...Option) (*http.ServeMux, error) {
	// configure gateway defaults
	g := gateway{
		serverAddress:     DefaultServerAddress,
		endpoints:         make(map[string][]registerFunc),
		serverDialOptions: []grpc.DialOption{grpc.WithInsecure()},
		mux:               http.NewServeMux(),
	}
	// apply functional options
	for _, opt := range options {
		opt(&g)
	}
	return g.registerEndpoints()
}

// registerEndpoints iterates through each prefix and registers its handlers
// to the REST gateway
func (g gateway) registerEndpoints() (*http.ServeMux, error) {
	for prefix, registers := range g.endpoints {
		gwmux := runtime.NewServeMux(
			append([]runtime.ServeMuxOption{runtime.WithProtoErrorHandler(ProtoMessageErrorHandler),
				runtime.WithMetadata(MetadataAnnotator)}, g.gatewayMuxOptions...)...,
		)
		for _, register := range registers {
			if err := register(
				context.Background(), gwmux, g.serverAddress, g.serverDialOptions,
			); err != nil {
				return nil, err
			}
		}
		// strip prefix from testRequest URI, but leave the trailing "/"
		g.mux.Handle(prefix, http.StripPrefix(prefix[:len(prefix)-1], gwmux))
	}
	return g.mux, nil
}

// WithDialOptions assigns a list of gRPC dial options to the REST gateway
func WithDialOptions(options ...grpc.DialOption) Option {
	return func(g *gateway) {
		g.serverDialOptions = options
	}
}

// WithEndpointRegistration takes a group of HTTP handlers that have been
// generated by the gRPC gateway protoc plugin and registers them to the REST
// gateway with some prefix (e.g. www.website.com/prefix/endpoint)
func WithEndpointRegistration(prefix string, endpoints ...registerFunc) Option {
	return func(g *gateway) {
		g.endpoints[prefix] = append(g.endpoints[prefix], endpoints...)
	}
}

// WithServerAddress determines what address the gateway will connect to. By
// default, the gateway will connect to 0.0.0.0:9090
func WithServerAddress(address string) Option {
	return func(g *gateway) {
		g.serverAddress = address
	}
}

// WithMux will use the given http.ServeMux to register the gateway endpoints.
func WithMux(mux *http.ServeMux) Option {
	return func(g *gateway) {
		g.mux = mux
	}
}

// WithGatewayOptions allows for additional gateway ServeMuxOptions beyond the
// default ProtoMessageErrorHandler and MetadataAnnotator from this package
func WithGatewayOptions(opt ...runtime.ServeMuxOption) Option {
	return func(g *gateway) {
		g.gatewayMuxOptions = append(g.gatewayMuxOptions, opt...)
	}
}
