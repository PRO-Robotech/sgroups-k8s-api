package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"google.golang.org/grpc"

	"sgroups.io/sgroups-k8s-api/internal/backend"
	"sgroups.io/sgroups-k8s-api/internal/grpcclient"
	"sgroups.io/sgroups-k8s-api/internal/mock"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

const (
	aggregatedServiceName      = "sgroups-apiserver"
	aggregatedServiceNamespace = "default"
)

func TestKubectlCreateContract(t *testing.T) {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl binary is required for contract test")
	}
	assetsDir := os.Getenv("KUBEBUILDER_ASSETS")
	if assetsDir == "" {
		t.Skip("KUBEBUILDER_ASSETS is not set; envtest binaries are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	env := &envtest.Environment{BinaryAssetsDirectory: assetsDir}
	cfg, err := env.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, env.Stop())
	})

	grpcAddr, grpcStop := startMockGRPCBackend(t)
	t.Cleanup(grpcStop)

	apiserverPort, caBundle, stopAPIServer := startAggregatedAPIServer(t, grpcAddr)
	t.Cleanup(stopAPIServer)

	require.NoError(t, registerAggregatedAPI(ctx, cfg, apiserverPort, caBundle))

	kubeconfigPath := writeKubeconfig(t, cfg)

	cases := []struct {
		name              string
		manifest          string
		getArgs           []string
		expectedName      string
		expectedNamespace string
	}{
		{
			name: "address-group",
			manifest: `apiVersion: sgroups.io/v1alpha1
kind: AddressGroup
metadata:
  name: ag-1
  namespace: default
spec:
  defaultAction: ALLOW
`,
			getArgs:           []string{"-n", "default", "addressgroups.sgroups.io", "ag-1"},
			expectedName:      "ag-1",
			expectedNamespace: "default",
		},
		{
			name: "tenant",
			manifest: `apiVersion: sgroups.io/v1alpha1
kind: Tenant
metadata:
  name: sg-default
spec:
  displayName: Default
`,
			getArgs:      []string{"tenants.sgroups.io", "sg-default"},
			expectedName: "sg-default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifestPath := writeManifest(t, tc.manifest)
			runKubectl(t, ctx, kubectlPath, kubeconfigPath, "create", "-f", manifestPath)

			getArgs := append([]string{"get"}, tc.getArgs...)
			getArgs = append(getArgs, "-o", "json")
			output := runKubectl(t, ctx, kubectlPath, kubeconfigPath, getArgs...)

			var obj map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &obj))
			metadata, _ := obj["metadata"].(map[string]any)
			require.Equal(t, tc.expectedName, metadata["name"])
			if tc.expectedNamespace != "" {
				require.Equal(t, tc.expectedNamespace, metadata["namespace"])
			}
		})
	}
}

func startMockGRPCBackend(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	mb := mock.New()
	b := backend.Backend{Namespaces: mb, AddressGroups: mb}
	mock.RegisterServices(grpcServer, b)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	stop := func() {
		grpcServer.GracefulStop()
		_ = listener.Close()
	}

	return listener.Addr().String(), stop
}

func startAggregatedAPIServer(t *testing.T, grpcAddr string) (int, []byte, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dnsName := fmt.Sprintf("%s.%s.svc", aggregatedServiceName, aggregatedServiceNamespace)
	altDNS := []string{
		"localhost",
		fmt.Sprintf("%s.%s.svc.cluster.local", aggregatedServiceName, aggregatedServiceNamespace),
	}
	certPEM, keyPEM, err := certutil.GenerateSelfSignedCertKey(dnsName, []net.IP{net.ParseIP("127.0.0.1")}, altDNS)
	require.NoError(t, err)

	certProvider, err := dynamiccertificates.NewStaticCertKeyContent("aggregated-apiserver", certPEM, keyPEM)
	require.NoError(t, err)

	opts := NewOptions()
	opts.Recommended.SecureServing.Listener = listener
	opts.Recommended.SecureServing.ServerCert.GeneratedCert = certProvider
	opts.Recommended.Authentication.RemoteKubeConfigFileOptional = true
	opts.Recommended.Authorization = nil
	opts.Recommended.CoreAPI = nil

	grpcClient, err := grpcclient.Dial(grpcclient.Config{
		Address:  grpcAddr,
		Insecure: true,
	})
	require.NoError(t, err)

	cfg, err := NewConfig(opts, grpcClient)
	require.NoError(t, err)

	srv, err := cfg.Complete().New()
	require.NoError(t, err)

	serverCtx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- srv.GenericAPIServer.PrepareRun().RunWithContext(serverCtx)
	}()

	httpClient, err := httpsClientWithCA(certPEM)
	require.NoError(t, err)
	require.NoError(t, waitForReadyz(serverCtx, httpClient, listener.Addr().(*net.TCPAddr).Port))

	stop := func() {
		cancel()
		_ = grpcClient.Close()
		_ = listener.Close()
		select {
		case <-runErr:
		case <-time.After(5 * time.Second):
		}
	}

	return listener.Addr().(*net.TCPAddr).Port, certPEM, stop
}

func registerAggregatedAPI(ctx context.Context, cfg *rest.Config, port int, caBundle []byte) error {
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aggregatedServiceName,
			Namespace: aggregatedServiceNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(port),
					TargetPort: intstr.FromInt(port),
				},
			},
		},
	}
	if _, err := kubeClient.CoreV1().Services(aggregatedServiceNamespace).Create(ctx, service, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	endpoints := &corev1.Endpoints{ //nolint:staticcheck // SA1019: Endpoints API needed for envtest compatibility
		ObjectMeta: metav1.ObjectMeta{
			Name:      aggregatedServiceName,
			Namespace: aggregatedServiceNamespace,
		},
		Subsets: []corev1.EndpointSubset{ //nolint:staticcheck // SA1019: part of deprecated Endpoints API
			{
				Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
				Ports: []corev1.EndpointPort{
					{
						Name:     "https",
						Protocol: corev1.ProtocolTCP,
						Port:     int32(port),
					},
				},
			},
		},
	}
	if _, err := kubeClient.CoreV1().Endpoints(aggregatedServiceNamespace).Create(ctx, endpoints, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	apiServiceName := fmt.Sprintf("%s.%s", v1alpha1.SchemeGroupVersion.Version, v1alpha1.GroupName)
	apiService := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apiregistration.k8s.io/v1",
			"kind":       "APIService",
			"metadata": map[string]any{
				"name": apiServiceName,
			},
			"spec": map[string]any{
				"group":                v1alpha1.GroupName,
				"version":              v1alpha1.SchemeGroupVersion.Version,
				"groupPriorityMinimum": int64(1000),
				"versionPriority":      int64(15),
				"service": map[string]any{
					"name":      aggregatedServiceName,
					"namespace": aggregatedServiceNamespace,
					"port":      int64(port),
				},
				"caBundle": base64.StdEncoding.EncodeToString(caBundle),
			},
		},
	}

	apiServiceGVR := schema.GroupVersionResource{
		Group:    "apiregistration.k8s.io",
		Version:  "v1",
		Resource: "apiservices",
	}
	if _, err := dynamicClient.Resource(apiServiceGVR).Create(ctx, apiService, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}

	return waitForGroupVersion(ctx, discoveryClient, v1alpha1.SchemeGroupVersion.String())
}

func waitForGroupVersion(ctx context.Context, client discovery.DiscoveryInterface, gv string) error {
	return wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, 20*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := client.ServerResourcesForGroupVersion(gv)
		if err != nil {
			return false, nil
		}

		return true, nil
	})
}

func waitForReadyz(ctx context.Context, client *http.Client, port int) error {
	url := fmt.Sprintf("https://127.0.0.1:%d/readyz", port)

	return wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, 10*time.Second, true, func(ctx context.Context) (bool, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return false, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, nil
		}
		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK, nil
	})
}

func httpsClientWithCA(caBundle []byte) (*http.Client, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBundle) {
		return nil, errors.New("failed to parse CA bundle")
	}
	tlsConfig := &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}

	return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}, nil
}

func writeKubeconfig(t *testing.T, cfg *rest.Config) string {
	t.Helper()

	clusterName := "envtest"
	userName := "envtest-user"
	contextName := "envtest"

	caData := firstNonEmpty(cfg.CAData, readFileIfSet(t, cfg.CAFile))
	certData := firstNonEmpty(cfg.CertData, readFileIfSet(t, cfg.CertFile))
	keyData := firstNonEmpty(cfg.KeyData, readFileIfSet(t, cfg.KeyFile))

	kubeconfig := clientcmdapi.NewConfig()
	kubeconfig.Clusters[clusterName] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: caData,
		InsecureSkipTLSVerify:    cfg.Insecure,
	}
	kubeconfig.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		ClientCertificateData: certData,
		ClientKeyData:         keyData,
		Token:                 cfg.BearerToken,
	}
	kubeconfig.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:  clusterName,
		AuthInfo: userName,
	}
	kubeconfig.CurrentContext = contextName

	path := filepath.Join(t.TempDir(), "kubeconfig")
	require.NoError(t, clientcmd.WriteToFile(*kubeconfig, path))

	return path
}

func writeManifest(t *testing.T, manifest string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "object.yaml")
	require.NoError(t, os.WriteFile(path, []byte(manifest), 0o600))

	return path
}

func runKubectl(t *testing.T, ctx context.Context, kubectlPath, kubeconfigPath string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(ctx, kubectlPath, append([]string{"--kubeconfig", kubeconfigPath}, args...)...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "kubectl %s failed: %s", strings.Join(args, " "), string(output))

	return string(output)
}

func readFileIfSet(t *testing.T, path string) []byte {
	t.Helper()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return data
}

func firstNonEmpty(primary []byte, fallback []byte) []byte {
	if len(primary) > 0 {
		return primary
	}

	return fallback
}
