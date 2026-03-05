package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFile_Full(t *testing.T) {
	yaml := `
serving:
  securePort: 9443
  tlsCertFile: /tmp/cert.pem
  tlsPrivateKeyFile: /tmp/key.pem

grpc:
  address: "backend:8081"
  insecure: true
  timeout: 45s
  ca: /tmp/ca.pem
  cert: /tmp/grpc-cert.pem
  key: /tmp/grpc-key.pem
  serverName: my-server
  maxRecvMsgSize: 33554432
  keepaliveTime: 60s
  keepaliveTimeout: 20s
  permitWithoutStream: false
  minConnectTimeout: 15s
  backoffMaxDelay: 240s
`
	fc := loadFromString(t, yaml)

	// Serving
	if fc.Serving == nil {
		t.Fatal("serving is nil")
	}
	assertIntPtr(t, "securePort", fc.Serving.SecurePort, 9443)
	assertStringPtr(t, "tlsCertFile", fc.Serving.TLSCertFile, "/tmp/cert.pem")
	assertStringPtr(t, "tlsPrivateKeyFile", fc.Serving.TLSPrivateKeyFile, "/tmp/key.pem")

	// GRPC
	if fc.GRPC == nil {
		t.Fatal("grpc is nil")
	}
	assertStringPtr(t, "address", fc.GRPC.Address, "backend:8081")
	assertBoolPtr(t, "insecure", fc.GRPC.Insecure, true)
	assertDurationPtr(t, "timeout", fc.GRPC.Timeout, 45*time.Second)
	assertStringPtr(t, "ca", fc.GRPC.CA, "/tmp/ca.pem")
	assertStringPtr(t, "cert", fc.GRPC.Cert, "/tmp/grpc-cert.pem")
	assertStringPtr(t, "key", fc.GRPC.Key, "/tmp/grpc-key.pem")
	assertStringPtr(t, "serverName", fc.GRPC.ServerName, "my-server")
	assertIntPtr(t, "maxRecvMsgSize", fc.GRPC.MaxRecvMsgSize, 33554432)
	assertDurationPtr(t, "keepaliveTime", fc.GRPC.KeepaliveTime, 60*time.Second)
	assertDurationPtr(t, "keepaliveTimeout", fc.GRPC.KeepaliveTimeout, 20*time.Second)
	assertBoolPtr(t, "permitWithoutStream", fc.GRPC.PermitWithoutStream, false)
	assertDurationPtr(t, "minConnectTimeout", fc.GRPC.MinConnectTimeout, 15*time.Second)
	assertDurationPtr(t, "backoffMaxDelay", fc.GRPC.BackoffMaxDelay, 240*time.Second)
}

func TestLoadFile_Partial(t *testing.T) {
	yaml := `
grpc:
  address: "backend:9090"
  insecure: true
`
	fc := loadFromString(t, yaml)

	if fc.Serving != nil {
		t.Error("serving should be nil for partial config")
	}
	if fc.GRPC == nil {
		t.Fatal("grpc is nil")
	}
	assertStringPtr(t, "address", fc.GRPC.Address, "backend:9090")
	assertBoolPtr(t, "insecure", fc.GRPC.Insecure, true)

	// Unset fields should be nil
	if fc.GRPC.Timeout != nil {
		t.Error("timeout should be nil")
	}
	if fc.GRPC.CA != nil {
		t.Error("ca should be nil")
	}
	if fc.GRPC.MaxRecvMsgSize != nil {
		t.Error("maxRecvMsgSize should be nil")
	}
}

func TestLoadFile_Empty(t *testing.T) {
	fc := loadFromString(t, "")

	if fc.Serving != nil {
		t.Error("serving should be nil for empty config")
	}
	if fc.GRPC != nil {
		t.Error("grpc should be nil for empty config")
	}
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	path := writeTemp(t, "not: valid: yaml: [")
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFile_InvalidDuration(t *testing.T) {
	yaml := `
grpc:
  timeout: "not-a-duration"
`
	path := writeTemp(t, yaml)
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestApplyFileConfig_NilFields(t *testing.T) {
	opts := NewOptions()
	origAddress := opts.GRPC.Address
	origTimeout := opts.GRPC.Timeout
	origPort := opts.Recommended.SecureServing.BindPort

	// Apply config with only address set — everything else should stay default.
	addr := "custom:1234"
	fc := &FileConfig{
		GRPC: &GRPCConfig{
			Address: &addr,
		},
	}
	opts.ApplyFileConfig(fc)

	if opts.GRPC.Address != "custom:1234" {
		t.Errorf("address: got %q, want %q", opts.GRPC.Address, "custom:1234")
	}
	if opts.GRPC.Timeout != origTimeout {
		t.Errorf("timeout changed: got %v, want %v", opts.GRPC.Timeout, origTimeout)
	}
	if opts.Recommended.SecureServing.BindPort != origPort {
		t.Errorf("port changed: got %d, want %d", opts.Recommended.SecureServing.BindPort, origPort)
	}
	_ = origAddress // suppress unused
}

func TestApplyFileConfig_AllFields(t *testing.T) {
	opts := NewOptions()

	port := 9443
	addr := "backend:8081"
	insecure := true
	timeout := Duration(45 * time.Second)
	ca := "/ca.pem"
	cert := "/cert.pem"
	key := "/key.pem"
	serverName := "srv"
	maxSize := 33554432
	keepTime := Duration(60 * time.Second)
	keepTimeout := Duration(20 * time.Second)
	permit := false
	minConnect := Duration(15 * time.Second)
	backoff := Duration(240 * time.Second)

	fc := &FileConfig{
		Serving: &ServingConfig{
			SecurePort: &port,
		},
		GRPC: &GRPCConfig{
			Address:             &addr,
			Insecure:            &insecure,
			Timeout:             &timeout,
			CA:                  &ca,
			Cert:                &cert,
			Key:                 &key,
			ServerName:          &serverName,
			MaxRecvMsgSize:      &maxSize,
			KeepaliveTime:       &keepTime,
			KeepaliveTimeout:    &keepTimeout,
			PermitWithoutStream: &permit,
			MinConnectTimeout:   &minConnect,
			BackoffMaxDelay:     &backoff,
		},
	}
	opts.ApplyFileConfig(fc)

	if opts.Recommended.SecureServing.BindPort != 9443 {
		t.Errorf("securePort: got %d, want 9443", opts.Recommended.SecureServing.BindPort)
	}
	if opts.GRPC.Address != "backend:8081" {
		t.Errorf("address: got %q", opts.GRPC.Address)
	}
	if opts.GRPC.Insecure != true {
		t.Error("insecure: got false")
	}
	if opts.GRPC.Timeout != 45*time.Second {
		t.Errorf("timeout: got %v", opts.GRPC.Timeout)
	}
	if opts.GRPC.CAFile != "/ca.pem" {
		t.Errorf("ca: got %q", opts.GRPC.CAFile)
	}
	if opts.GRPC.CertFile != "/cert.pem" {
		t.Errorf("cert: got %q", opts.GRPC.CertFile)
	}
	if opts.GRPC.KeyFile != "/key.pem" {
		t.Errorf("key: got %q", opts.GRPC.KeyFile)
	}
	if opts.GRPC.ServerName != "srv" {
		t.Errorf("serverName: got %q", opts.GRPC.ServerName)
	}
	if opts.GRPC.MaxRecvMsgSize != 33554432 {
		t.Errorf("maxRecvMsgSize: got %d", opts.GRPC.MaxRecvMsgSize)
	}
	if opts.GRPC.KeepaliveTime != 60*time.Second {
		t.Errorf("keepaliveTime: got %v", opts.GRPC.KeepaliveTime)
	}
	if opts.GRPC.KeepaliveTimeout != 20*time.Second {
		t.Errorf("keepaliveTimeout: got %v", opts.GRPC.KeepaliveTimeout)
	}
	if opts.GRPC.PermitWithoutStream != false {
		t.Error("permitWithoutStream: got true")
	}
	if opts.GRPC.MinConnectTimeout != 15*time.Second {
		t.Errorf("minConnectTimeout: got %v", opts.GRPC.MinConnectTimeout)
	}
	if opts.GRPC.BackoffMaxDelay != 240*time.Second {
		t.Errorf("backoffMaxDelay: got %v", opts.GRPC.BackoffMaxDelay)
	}
}

func TestApplyFileConfig_NilSections(t *testing.T) {
	opts := NewOptions()
	origAddress := opts.GRPC.Address
	origPort := opts.Recommended.SecureServing.BindPort

	opts.ApplyFileConfig(&FileConfig{})

	if opts.GRPC.Address != origAddress {
		t.Errorf("address changed: got %q, want %q", opts.GRPC.Address, origAddress)
	}
	if opts.Recommended.SecureServing.BindPort != origPort {
		t.Errorf("port changed: got %d, want %d", opts.Recommended.SecureServing.BindPort, origPort)
	}
}

// helpers

func loadFromString(t *testing.T, content string) *FileConfig {
	t.Helper()
	path := writeTemp(t, content)
	fc, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	return fc
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}

func assertStringPtr(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got nil, want %q", name, want)

		return
	}
	if *got != want {
		t.Errorf("%s: got %q, want %q", name, *got, want)
	}
}

func assertIntPtr(t *testing.T, name string, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got nil, want %d", name, want)

		return
	}
	if *got != want {
		t.Errorf("%s: got %d, want %d", name, *got, want)
	}
}

func assertBoolPtr(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got nil, want %v", name, want)

		return
	}
	if *got != want {
		t.Errorf("%s: got %v, want %v", name, *got, want)
	}
}

func assertDurationPtr(t *testing.T, name string, got *Duration, want time.Duration) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got nil, want %v", name, want)

		return
	}
	if got.Std() != want {
		t.Errorf("%s: got %v, want %v", name, got.Std(), want)
	}
}
