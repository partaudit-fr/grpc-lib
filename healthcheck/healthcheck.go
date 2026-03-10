package healthcheck

import (
	"context"

	"github.com/partaudit-fr/grpc-lib/protogen/go/health"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Info holds version and service name for the health check response.
// Provide this via fx to inject version info into the health endpoint.
type Info struct {
	Version string
	Service string
}

type healthCheck struct {
	health.HealthServiceServer
	info *Info
}

func newHealthCheck(params healthCheckParams) health.HealthServiceServer {
	info := params.Info
	if info == nil {
		info = &Info{Version: "unknown", Service: "unknown"}
	}
	return &healthCheck{info: info}
}

func (h *healthCheck) Check(_ context.Context, _ *emptypb.Empty) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status:  "ok",
		Version: h.info.Version,
		Service: h.info.Service,
	}, nil
}
