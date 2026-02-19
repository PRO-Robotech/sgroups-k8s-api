package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the YAML configuration file structure.
// All fields are pointers so that absent YAML keys are distinguishable from
// zero values and do not override defaults.
type FileConfig struct {
	Serving *ServingConfig `yaml:"serving,omitempty"`
	GRPC    *GRPCConfig    `yaml:"grpc,omitempty"`
}

// ServingConfig holds serving-related settings from the YAML file.
type ServingConfig struct {
	SecurePort        *int    `yaml:"securePort,omitempty"`
	TLSCertFile       *string `yaml:"tlsCertFile,omitempty"`
	TLSPrivateKeyFile *string `yaml:"tlsPrivateKeyFile,omitempty"`
}

// GRPCConfig holds gRPC client settings from the YAML file.
type GRPCConfig struct {
	Address             *string   `yaml:"address,omitempty"`
	Insecure            *bool     `yaml:"insecure,omitempty"`
	Timeout             *Duration `yaml:"timeout,omitempty"`
	CA                  *string   `yaml:"ca,omitempty"`
	Cert                *string   `yaml:"cert,omitempty"`
	Key                 *string   `yaml:"key,omitempty"`
	ServerName          *string   `yaml:"serverName,omitempty"`
	MaxRecvMsgSize      *int      `yaml:"maxRecvMsgSize,omitempty"`
	KeepaliveTime       *Duration `yaml:"keepaliveTime,omitempty"`
	KeepaliveTimeout    *Duration `yaml:"keepaliveTimeout,omitempty"`
	PermitWithoutStream *bool     `yaml:"permitWithoutStream,omitempty"`
	MinConnectTimeout   *Duration `yaml:"minConnectTimeout,omitempty"`
	BackoffMaxDelay     *Duration `yaml:"backoffMaxDelay,omitempty"`
}

// Duration wraps time.Duration to support YAML string unmarshalling
// (e.g. "30s", "5m", "1h30m").
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)

	return nil
}

// Std returns the underlying time.Duration.
func (d *Duration) Std() time.Duration {
	return time.Duration(*d)
}

// LoadFile reads and parses a YAML configuration file at the given path.
func LoadFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is an explicit user-provided config file path
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &fc, nil
}
