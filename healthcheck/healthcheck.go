package healthcheck

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/partaudit-fr/grpc-lib/protogen/go/health"
)

type healthCheck struct {
	health.HealthServiceServer
}

func newHealthCheck() health.HealthServiceServer {
	return &healthCheck{}
}

func (h *healthCheck) Check(_ context.Context, _ *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
