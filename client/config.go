package client

type GRPCConfigClient interface {
	Port() int
	ServiceName() string
}

type YAMLGRPCConfigClient struct {
	ValuePort        int    `yaml:"port"`
	ValueServiceName string `yaml:"service_name"`
}

func (c YAMLGRPCConfigClient) Port() int {
	return c.ValuePort
}

func (c YAMLGRPCConfigClient) ServiceName() string {
	if c.ValueServiceName == "" {
		return "http-gateway" // fallback par défaut
	}
	return c.ValueServiceName
}
