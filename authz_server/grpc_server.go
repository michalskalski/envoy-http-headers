package main

/*
  This is close variation of jbarratt@ repo here
  https://github.com/jbarratt/envoy_ratelimit_example/blob/master/extauth/main.go

*/
import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
)

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

type AuthorizationServer struct{}

func (a *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	action, ok := req.Attributes.Request.Http.Headers["action"]

	if ok {
		switch action {
		case "add-if-absent":
			return &auth.CheckResponse{
				Status: &rpcstatus.Status{
					Code: int32(rpc.OK),
				},
				HttpResponse: &auth.CheckResponse_OkResponse{
					OkResponse: &auth.OkHttpResponse{
						ResponseHeadersToAdd: []*core.HeaderValueOption{
							{
								Header: &core.HeaderValue{
									Key:   "test-header",
									Value: "add-if-absent",
								},
								AppendAction: core.HeaderValueOption_ADD_IF_ABSENT,
							},
						},
					},
				},
			}, nil
		case "append-if-exist":
			return &auth.CheckResponse{
				Status: &rpcstatus.Status{
					Code: int32(rpc.OK),
				},
				HttpResponse: &auth.CheckResponse_OkResponse{
					OkResponse: &auth.OkHttpResponse{
						ResponseHeadersToAdd: []*core.HeaderValueOption{
							{
								Header: &core.HeaderValue{
									Key:   "test-header",
									Value: "append-if-exist",
								},
								AppendAction: core.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
							},
						},
					},
				},
			}, nil
		default:
			return &auth.CheckResponse{
				Status: &rpcstatus.Status{
					Code: int32(rpc.UNAUTHENTICATED),
				},
				HttpResponse: &auth.CheckResponse_DeniedResponse{
					DeniedResponse: &auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: "Unknow action",
					},
				},
			}, nil
		}
	} else {
		return &auth.CheckResponse{
			Status: &rpcstatus.Status{
				Code: int32(rpc.UNAUTHENTICATED),
			},
			HttpResponse: &auth.CheckResponse_DeniedResponse{
				DeniedResponse: &auth.DeniedHttpResponse{
					Status: &envoy_type.HttpStatus{
						Code: envoy_type.StatusCode_Unauthorized,
					},
					Body: "Action header not provided",
				},
			},
		}, nil

	}
}

func main() {
	address := ":5001"
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	auth.RegisterAuthorizationServer(g, &AuthorizationServer{})
	healthpb.RegisterHealthServer(g, &healthServer{})
	log.Printf("gRPC server listen at %s", address)
	g.Serve(l)

}
