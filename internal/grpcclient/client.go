package grpcclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"sgroups.io/sgroups-k8s-api/pkg/client"
)

// Config configures the gRPC client connection.
type Config struct {
	Address             string
	CAFile              string
	CertFile            string
	KeyFile             string
	ServerName          string
	Insecure            bool
	MaxRecvMsgSize      int
	KeepaliveTime       time.Duration
	KeepaliveTimeout    time.Duration
	PermitWithoutStream bool
	MinConnectTimeout   time.Duration
	BackoffMaxDelay     time.Duration
}

// Dial establishes a gRPC connection to the backend with optional mTLS.
func Dial(cfg Config, opts ...grpc.DialOption) (*client.Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("grpc address is required")
	}
	dialOpts := append([]grpc.DialOption{}, opts...)
	if len(opts) == 0 {
		creds, err := credentialsForConfig(cfg)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}
	dialOpts = append(dialOpts, dialOptions(cfg)...)
	dialOpts = append(dialOpts,
		grpc.WithChainUnaryInterceptor(UserMetadataUnaryInterceptor()),
		grpc.WithChainStreamInterceptor(UserMetadataStreamInterceptor()),
	)

	return client.Dial(cfg.Address, dialOpts...)
}

func credentialsForConfig(cfg Config) (credentials.TransportCredentials, error) {
	if cfg.Insecure {
		return insecure.NewCredentials(), nil
	}

	var certs []tls.Certificate
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, errors.New("both cert and key are required for mTLS")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		certs = []tls.Certificate{cert}
	}

	rootCAs, err := loadRootCAs(cfg.CAFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      rootCAs,
		Certificates: certs,
		ServerName:   cfg.ServerName,
	}

	return credentials.NewTLS(tlsCfg), nil
}

func loadRootCAs(caFile string) (*x509.CertPool, error) {
	if caFile == "" {
		return x509.SystemCertPool()
	}
	data, err := os.ReadFile(caFile) //nolint:gosec // G304: CA file path comes from CLI flags, not user input
	if err != nil {
		return nil, fmt.Errorf("read ca file: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, errors.New("failed to parse CA bundle")
	}

	return pool, nil
}

func dialOptions(cfg Config) []grpc.DialOption {
	var opts []grpc.DialOption
	if cfg.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize)))
	}
	if cfg.KeepaliveTime > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.KeepaliveTime,
			Timeout:             cfg.KeepaliveTimeout,
			PermitWithoutStream: cfg.PermitWithoutStream,
		}))
	}
	if cfg.MinConnectTimeout > 0 || cfg.BackoffMaxDelay > 0 {
		bcfg := backoff.DefaultConfig
		if cfg.BackoffMaxDelay > 0 {
			bcfg.MaxDelay = cfg.BackoffMaxDelay
		}
		connectTimeout := cfg.MinConnectTimeout
		if connectTimeout <= 0 {
			connectTimeout = 20 * time.Second
		}
		opts = append(opts, grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           bcfg,
			MinConnectTimeout: connectTimeout,
		}))
	}

	return opts
}
