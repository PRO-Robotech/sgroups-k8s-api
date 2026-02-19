package config

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server/options"

	"sgroups.io/sgroups-k8s-api/internal/grpcclient"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// Options configures the aggregated API server.
type Options struct {
	Recommended *options.RecommendedOptions
	GRPC        GRPCOptions
}

// GRPCOptions configures the gRPC client connection.
type GRPCOptions struct {
	Address             string
	CAFile              string
	CertFile            string
	KeyFile             string
	ServerName          string
	Insecure            bool
	Timeout             time.Duration
	MaxRecvMsgSize      int
	KeepaliveTime       time.Duration
	KeepaliveTimeout    time.Duration
	PermitWithoutStream bool
	MinConnectTimeout   time.Duration
	BackoffMaxDelay     time.Duration
}

// NewOptions returns options with defaults.
func NewOptions() *Options {
	recommended := options.NewRecommendedOptions(
		"/registry/sgroups.io",
		v1alpha1.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion),
	)
	recommended.Etcd = nil
	recommended.SecureServing.BindPort = 8443

	return &Options{
		Recommended: recommended,
		GRPC: GRPCOptions{
			Address:             "127.0.0.1:8081",
			Insecure:            false,
			Timeout:             30 * time.Second,
			MaxRecvMsgSize:      16 * 1024 * 1024,
			KeepaliveTime:       30 * time.Second,
			KeepaliveTimeout:    10 * time.Second,
			PermitWithoutStream: true,
			MinConnectTimeout:   10 * time.Second,
			BackoffMaxDelay:     120 * time.Second,
		},
	}
}

// AddFlags adds flags for options.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Recommended.AddFlags(fs)
	fs.StringVar(&o.GRPC.Address, "grpc-address", o.GRPC.Address, "SGroups gRPC backend address")
	fs.StringVar(&o.GRPC.CAFile, "grpc-ca", o.GRPC.CAFile, "CA bundle for gRPC TLS")
	fs.StringVar(&o.GRPC.CertFile, "grpc-cert", o.GRPC.CertFile, "Client cert for gRPC mTLS")
	fs.StringVar(&o.GRPC.KeyFile, "grpc-key", o.GRPC.KeyFile, "Client key for gRPC mTLS")
	fs.StringVar(&o.GRPC.ServerName, "grpc-server-name", o.GRPC.ServerName, "Override gRPC TLS server name")
	fs.BoolVar(&o.GRPC.Insecure, "grpc-insecure", o.GRPC.Insecure, "Disable TLS for gRPC (not for production)")
	fs.DurationVar(&o.GRPC.Timeout, "grpc-timeout", o.GRPC.Timeout, "Default timeout for gRPC requests (0 disables)")
	fs.IntVar(&o.GRPC.MaxRecvMsgSize, "grpc-max-recv-msg-size", o.GRPC.MaxRecvMsgSize,
		"Max gRPC receive message size in bytes (0 uses default)")
	fs.DurationVar(&o.GRPC.KeepaliveTime, "grpc-keepalive-time", o.GRPC.KeepaliveTime, "gRPC keepalive ping interval (0 disables)")
	fs.DurationVar(&o.GRPC.KeepaliveTimeout, "grpc-keepalive-timeout", o.GRPC.KeepaliveTimeout, "gRPC keepalive ping timeout")
	fs.BoolVar(&o.GRPC.PermitWithoutStream, "grpc-keepalive-permit-without-stream",
		o.GRPC.PermitWithoutStream, "Allow keepalive pings without active streams")
	fs.DurationVar(&o.GRPC.MinConnectTimeout, "grpc-min-connect-timeout", o.GRPC.MinConnectTimeout, "Minimum gRPC connect timeout")
	fs.DurationVar(&o.GRPC.BackoffMaxDelay, "grpc-backoff-max-delay", o.GRPC.BackoffMaxDelay, "Maximum gRPC backoff delay")
}

// Validate returns an error for invalid options.
func (o *Options) Validate() error {
	if o.GRPC.Address == "" {
		return errors.New("grpc-address is required")
	}

	return nil
}

// ApplyFileConfig overwrites option values with non-nil fields from the YAML
// config. Call this after NewOptions and before AddFlags so that YAML values
// become the pflag defaults, which CLI flags can then override.
//
// Serving TLS fields (TLSCertFile, TLSPrivateKeyFile) cannot be set here
// because SecureServingOptions registers them internally. Use
// ApplyServingOverrides after Parse for those.
func (o *Options) ApplyFileConfig(fc *FileConfig) {
	if fc.Serving != nil {
		o.applyServingConfig(fc.Serving)
	}
	if fc.GRPC != nil {
		o.applyGRPCConfig(fc.GRPC)
	}
}

func (o *Options) applyServingConfig(sc *ServingConfig) {
	if sc.SecurePort != nil {
		o.Recommended.SecureServing.BindPort = *sc.SecurePort
	}
}

func (o *Options) applyGRPCConfig(gc *GRPCConfig) { //nolint:gocyclo // flat nil-guards for each config field, not real complexity
	if gc.Address != nil {
		o.GRPC.Address = *gc.Address
	}
	if gc.Insecure != nil {
		o.GRPC.Insecure = *gc.Insecure
	}
	if gc.Timeout != nil {
		o.GRPC.Timeout = gc.Timeout.Std()
	}
	if gc.CA != nil {
		o.GRPC.CAFile = *gc.CA
	}
	if gc.Cert != nil {
		o.GRPC.CertFile = *gc.Cert
	}
	if gc.Key != nil {
		o.GRPC.KeyFile = *gc.Key
	}
	if gc.ServerName != nil {
		o.GRPC.ServerName = *gc.ServerName
	}
	if gc.MaxRecvMsgSize != nil {
		o.GRPC.MaxRecvMsgSize = *gc.MaxRecvMsgSize
	}
	if gc.KeepaliveTime != nil {
		o.GRPC.KeepaliveTime = gc.KeepaliveTime.Std()
	}
	if gc.KeepaliveTimeout != nil {
		o.GRPC.KeepaliveTimeout = gc.KeepaliveTimeout.Std()
	}
	if gc.PermitWithoutStream != nil {
		o.GRPC.PermitWithoutStream = *gc.PermitWithoutStream
	}
	if gc.MinConnectTimeout != nil {
		o.GRPC.MinConnectTimeout = gc.MinConnectTimeout.Std()
	}
	if gc.BackoffMaxDelay != nil {
		o.GRPC.BackoffMaxDelay = gc.BackoffMaxDelay.Std()
	}
}

// GRPCConfig converts GRPCOptions into a dial config.
func (o GRPCOptions) GRPCConfig() grpcclient.Config {
	return grpcclient.Config{
		Address:             o.Address,
		CAFile:              o.CAFile,
		CertFile:            o.CertFile,
		KeyFile:             o.KeyFile,
		ServerName:          o.ServerName,
		Insecure:            o.Insecure,
		MaxRecvMsgSize:      o.MaxRecvMsgSize,
		KeepaliveTime:       o.KeepaliveTime,
		KeepaliveTimeout:    o.KeepaliveTimeout,
		PermitWithoutStream: o.PermitWithoutStream,
		MinConnectTimeout:   o.MinConnectTimeout,
		BackoffMaxDelay:     o.BackoffMaxDelay,
	}
}
