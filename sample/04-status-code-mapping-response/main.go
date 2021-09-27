package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	v3alpha "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3alpha"
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

const (
	HTTP_METHOD int = iota
	RES_CONTENT int = iota
)

func (server *Server) Process(processServer pb.ExternalProcessor_ProcessServer) error {
	ctx := processServer.Context()

	htmlMap := map[string]string{
		"403": share.ReadFile("public/403.html"),
		"404": share.ReadFile("public/404.html"),
		"500": share.ReadFile("public/500.html"),
		"503": share.ReadFile("public/503.html"),
	}

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
			ctx = context.WithValue(ctx, HTTP_METHOD, httpMethod)
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
			responseStatus, _ := share.GetHeaderValue(value.ResponseHeaders.Headers.Headers, ":status")
			logrus.Print(fmt.Sprintf("Handle (REQ_HEAD): upstream -> ext_proc -> downstream, status:%v", responseStatus))
			httpMethod := (ctx).Value(HTTP_METHOD).(string)
			html := ""
			for status, resHtml := range htmlMap {
				if responseStatus == status {
					html = resHtml
				}
			}
			ctx = context.WithValue(ctx, RES_CONTENT, html)
			if httpMethod == "GET" && html != "" {
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_ResponseHeaders{
						ResponseHeaders: &pb.HeadersResponse{
							Response: &pb.CommonResponse{
								HeaderMutation: &pb.HeaderMutation{
									RemoveHeaders: []string{
										"content-encoding",
									},
									SetHeaders: []*core.HeaderValueOption{
										{
											Header: &core.HeaderValue{
												Key:   "content-type",
												Value: "text/html",
											},
										},
										{
											Header: &core.HeaderValue{
												Key:   "content-length",
												Value: strconv.Itoa(len(html)),
											},
										},
									},
								},
							},
						},
					},
					ModeOverride: &v3alpha.ProcessingMode{
						ResponseBodyMode: v3alpha.ProcessingMode_BUFFERED,
					},
				}
			} else {
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_ResponseHeaders{},
				}
			}
			break
		case *pb.ProcessingRequest_ResponseBody:
			logrus.Print("Handle (REQ_BODY): upstream -> ext_proc -> downstream")
			httpMethod := (ctx).Value(HTTP_METHOD).(string)
			html := (ctx).Value(RES_CONTENT).(string)
			if httpMethod == "GET" && html != "" {
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_ResponseBody{
						ResponseBody: &pb.BodyResponse{
							Response: &pb.CommonResponse{
								BodyMutation: &pb.BodyMutation{
									Mutation: &pb.BodyMutation_Body{
										Body: []byte(html),
									},
								},
							},
						},
					},
					ModeOverride: &v3alpha.ProcessingMode{},
				}
			} else {
				resp = &pb.ProcessingResponse{
					Response: &pb.ProcessingResponse_ResponseBody{},
				}
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
