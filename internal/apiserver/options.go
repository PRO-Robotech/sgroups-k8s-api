package apiserver

import "sgroups.io/sgroups-k8s-api/internal/config"

// Options is an alias for config.Options.
type Options = config.Options

// GRPCOptions is an alias for config.GRPCOptions.
type GRPCOptions = config.GRPCOptions

// NewOptions returns options with defaults.
func NewOptions() *Options {
	return config.NewOptions()
}
