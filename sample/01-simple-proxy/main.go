package main

import (
	"fmt"
	"io"

	extProc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3alpha"
	pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3alpha"
	"github.com/sirupsen/logrus"
	"github.comt/Himenon/envoy-ext-proc-sample/share"
	"google.golang.org/grpc/codes"
	grpcHealthPb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
}

func (server *Server) Process(processServer pb.ExternalProcessor_ProcessServer) error {
	ctx := processServer.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := processServer.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		resp := &pb.ProcessingResponse{}
		switch value := req.Request.(type) {
		case *pb.ProcessingRequest_RequestHeaders:
			httpMethod, _ := share.GetHeaderValue(value.RequestHeaders.Headers.Headers, ":method")
			requestPath, _ := share.GetHeaderValue(value.RequestHeaders.Headers.Headers, ":path")
			logrus.Print(fmt.Sprintf("Handle (REQ_HEAD): downstream -> ext_proc -> upstream, Method:%s, Path:%s", httpMethod, requestPath))
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_RequestHeaders{},
			}
			break
		case *pb.ProcessingRequest_RequestBody:
			logrus.Print("Handle (REQ_BODY): downstream -> ext_proc -> upstream")
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_RequestBody{},
			}
			break
		case *pb.ProcessingRequest_ResponseHeaders:
			status, _ := share.GetHeaderValue(value.ResponseHeaders.Headers.Headers, ":status")
			logrus.Print(fmt.Sprintf("Handle (REQ_HEAD): upstream -> ext_proc -> downstream, status:%v", status))
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_ResponseHeaders{},
			}
			break
		case *pb.ProcessingRequest_ResponseBody:
			logrus.Print("Handle (REQ_BODY): upstream -> ext_proc -> downstream")
			resp = &pb.ProcessingResponse{
				Response: &pb.ProcessingResponse_ResponseBody{},
			}
			break
		default:
			logrus.Debug(fmt.Sprintf("Unknown Request type %v\n", value))
		}
		if err := processServer.Send(resp); err != nil {
			logrus.Debug(fmt.Sprintf("send error %v", err))
		}
	}
}

func main() {
	GRPC_ADDRESS := share.GetEnv("GRPC_ADDRESS", ":18080")
	tcpListener := share.CreateTcpListener(GRPC_ADDRESS)
	grpcServer := share.CreateGrpcServer(1000)
	server := Server{}
	extProc.RegisterExternalProcessorServer(grpcServer, &server)
	grpcHealthPb.RegisterHealthServer(grpcServer, &share.HealthServer{})
	logrus.Info(fmt.Sprintf("Starting gRPC server on address %s", GRPC_ADDRESS))
	grpcServer.Serve(tcpListener)
}
