package apiserver

import (
	"errors"
	"net/http"

	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/util/compatibility"

	"sgroups.io/sgroups-k8s-api/internal/apiserver/filters"
	registryoptions "sgroups.io/sgroups-k8s-api/internal/registry/options"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
	"sgroups.io/sgroups-k8s-api/pkg/client"
)

// Config contains the server configuration.
type Config struct {
	GenericConfig *server.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// ExtraConfig contains custom configuration.
type ExtraConfig struct {
	GRPCClient     *client.Client
	StorageOptions registryoptions.StorageOptions
}

// NewConfig builds a config from options and gRPC client.
func NewConfig(opts *Options, grpcClient *client.Client) (*Config, error) {
	if opts == nil {
		return nil, errors.New("options are required")
	}
	if grpcClient == nil {
		return nil, errors.New("grpc client is required")
	}

	genericConfig := server.NewRecommendedConfig(v1alpha1.Codecs)

	// EffectiveVersion must be set before Complete() — RecommendedOptions
	// does not include ServerRunOptions which normally sets it.
	genericConfig.EffectiveVersion = compatibility.DefaultBuildEffectiveVersion()

	if err := opts.Recommended.ApplyTo(genericConfig); err != nil {
		return nil, err
	}

	genericConfig.OpenAPIConfig = server.DefaultOpenAPIConfig(
		v1alpha1.GetOpenAPIDefinitionsWithEnums,
		openapi.NewDefinitionNamer(v1alpha1.Scheme),
	)
	genericConfig.OpenAPIConfig.Info.Title = "SGroups"
	genericConfig.OpenAPIConfig.Info.Version = v1alpha1.SchemeGroupVersion.Version

	genericConfig.OpenAPIV3Config = server.DefaultOpenAPIV3Config(
		v1alpha1.GetOpenAPIDefinitionsWithEnums,
		openapi.NewDefinitionNamer(v1alpha1.Scheme),
	)
	genericConfig.OpenAPIV3Config.Info.Title = "SGroups"
	genericConfig.OpenAPIV3Config.Info.Version = v1alpha1.SchemeGroupVersion.Version

	genericConfig.BuildHandlerChainFunc = func(handler http.Handler, c *server.Config) http.Handler {
		handler = filters.RejectApplyPatch(v1alpha1.Codecs)(handler)
		handler = server.DefaultBuildHandlerChain(handler, c)
		handler = filters.StripExtraAuthHeaders(handler)

		return handler
	}

	return &Config{
		GenericConfig: genericConfig,
		ExtraConfig: ExtraConfig{
			GRPCClient:     grpcClient,
			StorageOptions: registryoptions.StorageOptions{Timeout: opts.GRPC.Timeout},
		},
	}, nil
}
