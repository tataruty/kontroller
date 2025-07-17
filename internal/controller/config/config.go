package config

import (
	"github.com/go-logr/logr"
)

type Config struct {
	// Logger is the Zap Logger used by all components.
	Logger logr.Logger
	// GatewayCtlrName is the name of this controller.
	GatewayCtlrName string
	// ConfigName is the name of the NginxGateway resource for this controller.
	ConfigName string
	// GatewayClassName is the name of the GatewayClass resource that the Gateway will use.
	GatewayClassName string
}
