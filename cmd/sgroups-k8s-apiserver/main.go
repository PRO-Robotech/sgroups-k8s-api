package main

import (
	"context"
	goflag "flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"sgroups.io/sgroups-k8s-api/internal/apiserver"
	"sgroups.io/sgroups-k8s-api/internal/config"
	"sgroups.io/sgroups-k8s-api/internal/grpcclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := apiserver.NewOptions()

	// 1. Pre-parse --config before full flag registration.
	configPath := preParseConfigFlag(os.Args[1:])
	var fc *config.FileConfig
	if configPath != "" {
		var err error
		fc, err = config.LoadFile(configPath)
		if err != nil {
			klog.Fatalf("load config: %v", err)
		}
		opts.ApplyFileConfig(fc)
	}

	// 2. Setup flags — YAML values are now the pflag defaults.
	klog.InitFlags(nil)
	fs := pflag.NewFlagSet("sgroups-k8s-apiserver", pflag.ExitOnError)
	fs.AddGoFlagSet(goflag.CommandLine) // bridge klog flags (e.g. --v) into pflag
	fs.String("config", "", "Path to YAML configuration file")
	opts.AddFlags(fs)

	// 3. Parse CLI — explicit flags override YAML values.
	if err := fs.Parse(os.Args[1:]); err != nil {
		klog.Fatalf("parse flags: %v", err)
	}

	// 4. Apply serving TLS fields from YAML that CLI did not override.
	if fc != nil {
		applyServingOverrides(fs, fc)
	}

	if err := opts.Validate(); err != nil {
		klog.Fatalf("invalid options: %v", err)
	}

	grpcClient, err := grpcclient.Dial(opts.GRPC.GRPCConfig())
	if err != nil {
		klog.Fatalf("grpc dial: %v", err)
	}
	defer func() {
		_ = grpcClient.Close()
	}()

	cfg, err := apiserver.NewConfig(opts, grpcClient)
	if err != nil {
		klog.Fatalf("config: %v", err)
	}

	srv, err := cfg.Complete().New()
	if err != nil {
		klog.Fatalf("server: %v", err)
	}

	if err := srv.GenericAPIServer.PrepareRun().RunWithContext(ctx); err != nil {
		klog.Fatalf("run: %v", err)
	}
}

// preParseConfigFlag scans args for --config=<path> or --config <path>
// without requiring full flag registration. Returns "" if not found.
func preParseConfigFlag(args []string) string {
	for i, arg := range args {
		if arg == "--" {
			return ""
		}
		if v, ok := strings.CutPrefix(arg, "--config="); ok {
			return v
		}
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

// applyServingOverrides sets pflag values for serving fields from the YAML
// config where the corresponding CLI flag was not explicitly provided.
// This is needed because SecureServingOptions registers tls-cert-file and
// tls-private-key-file internally, so we cannot set them before AddFlags.
func applyServingOverrides(fs *pflag.FlagSet, fc *config.FileConfig) {
	if fc.Serving == nil {
		return
	}
	setIfUnchanged(fs, "tls-cert-file", fc.Serving.TLSCertFile)
	setIfUnchanged(fs, "tls-private-key-file", fc.Serving.TLSPrivateKeyFile)
}

// setIfUnchanged sets a pflag value only when the YAML config provides it
// and the user did not explicitly pass the flag on the command line.
func setIfUnchanged(fs *pflag.FlagSet, name string, val *string) {
	if val == nil {
		return
	}
	if fs.Changed(name) {
		return
	}
	if err := fs.Set(name, *val); err != nil {
		klog.Fatalf("set flag %s: %v", name, err)
	}
}
